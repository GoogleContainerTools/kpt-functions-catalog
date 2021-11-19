// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	opaapis "github.com/open-policy-agent/frameworks/constraint/pkg/apis"
	"github.com/open-policy-agent/frameworks/constraint/pkg/apis/templates/v1"
	opaclient "github.com/open-policy-agent/frameworks/constraint/pkg/client"
	"github.com/open-policy-agent/frameworks/constraint/pkg/client/drivers/local"
	"github.com/open-policy-agent/frameworks/constraint/pkg/core/templates"
	opatypes "github.com/open-policy-agent/frameworks/constraint/pkg/types"
	"github.com/open-policy-agent/gatekeeper/pkg/target"
	opautil "github.com/open-policy-agent/gatekeeper/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var scheme = runtime.NewScheme()

// All versions (v1alpha1, v1beta1 and v1) of the ConstraintTemplate have implemented this interface.
// We use it as suggested in https://github.com/GoogleContainerTools/kpt-functions-catalog/pull/649#discussion_r752553926.
type versionless interface {
	ToVersionless() (*templates.ConstraintTemplate, error)
}

func init() {
	err := opaapis.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
}

func createClient() (*opaclient.Client, error) {
	driver := local.New(local.Tracing(false))
	backend, err := opaclient.NewBackend(opaclient.Driver(driver))
	if err != nil {
		return nil, err
	}
	return backend.NewClient(opaclient.Targets(&target.K8sValidationTarget{}))
}

func gatherTemplates(objects []runtime.Object) ([]*templates.ConstraintTemplate, error) {
	var templs []*templates.ConstraintTemplate
	for _, obj := range objects {
		gvk := obj.GetObjectKind().GroupVersionKind()
		if gvk.Group != v1.SchemeGroupVersion.Group || gvk.Kind != "ConstraintTemplate" {
			continue
		}
		v, isVersionless := obj.(versionless)
		if !isVersionless {
			continue
		}
		template, err := v.ToVersionless()
		if err != nil {
			return nil, fmt.Errorf("unable to convert the ConstraintTemplate: %w", err)
		}
		templs = append(templs, template)
	}
	return templs, nil
}

func gatherConstraints(objects []runtime.Object) ([]*unstructured.Unstructured, error) {
	var cstrs []*unstructured.Unstructured
	for _, obj := range objects {
		gvk := obj.GetObjectKind().GroupVersionKind()
		if gvk.Group == "constraints.gatekeeper.sh" {
			cstrs = append(cstrs, obj.(*unstructured.Unstructured))
		}
	}
	return cstrs, nil
}

// Validate makes sure the configs passed to it comply with any Constraints and
// Constraint Templates present in the list of configs
func Validate(objects []runtime.Object) (*framework.Result, error) {
	client, err := createClient()
	if err != nil {
		return nil, err
	}
	tmpls, err := gatherTemplates(objects)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	for _, t := range tmpls {
		if _, err = client.AddTemplate(ctx, t); err != nil {
			return nil, err
		}
	}
	cstrs, err := gatherConstraints(objects)
	if err != nil {
		return nil, err
	}
	for _, c := range cstrs {
		if _, err = client.AddConstraint(ctx, c); err != nil {
			return nil, err
		}
	}

	for _, obj := range objects {
		if _, err = client.AddData(ctx, obj); err != nil {
			return nil, err
		}
	}

	resps, err := client.Audit(ctx)
	if err != nil {
		return nil, err
	}
	results := resps.Results()
	if len(results) > 0 {
		return parseResults(results)
	}
	return nil, nil
}

func parseResults(results []*opatypes.Result) (*framework.Result, error) {
	var items []framework.ResultItem

	for _, r := range results {
		u, ok := r.Resource.(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("could not cast to unstructured: %+v", r.Resource)
		}

		item := framework.ResultItem{
			Message: fmt.Sprintf("%s\nviolatedConstraint: %s", r.Msg, r.Constraint.GetName()),
			ResourceRef: yaml.ResourceIdentifier{
				TypeMeta: yaml.TypeMeta{
					APIVersion: u.GetAPIVersion(),
					Kind:       u.GetKind(),
				},
				NameMeta: yaml.NameMeta{
					Name:      u.GetName(),
					Namespace: u.GetNamespace(),
				},
			},
		}

		switch r.EnforcementAction {
		case string(opautil.Dryrun):
			item.Severity = framework.Info
		case string(opautil.Warn):
			item.Severity = framework.Warning
		default:
			item.Severity = framework.Error
		}

		path, foundPath := u.GetAnnotations()[kioutil.PathAnnotation]
		index, foundIndex := u.GetAnnotations()[kioutil.IndexAnnotation]
		if foundPath {
			item.File = framework.File{
				Path: path,
			}
			if foundIndex {
				idx, err := strconv.Atoi(index)
				if err != nil {
					return nil, err
				}
				item.File.Index = idx
			}
		}

		items = append(items, item)
	}
	sortResultItems(items)

	return &framework.Result{
		Items: items,
	}, nil
}

// TODO(mengqiy): upstream this to the SDK
func sortResultItems(items []framework.ResultItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if fileLess(items, i, j) != 0 {
			return fileLess(items, i, j) < 0
		}
		if severityLess(items, i, j) != 0 {
			return severityLess(items, i, j) < 0
		}
		return resultItemToString(items[i]) < resultItemToString(items[j])
	})
}

func severityLess(items []framework.ResultItem, i, j int) int {
	severityToNumber := map[framework.Severity]int{
		framework.Error:   0,
		framework.Warning: 1,
		framework.Info:    2,
	}

	severityLevelI, found := severityToNumber[items[i].Severity]
	if !found {
		severityLevelI = 3
	}
	severityLevelJ, found := severityToNumber[items[j].Severity]
	if !found {
		severityLevelJ = 3
	}
	return severityLevelI - severityLevelJ
}

func fileLess(items []framework.ResultItem, i, j int) int {
	if items[i].File.Path != items[j].File.Path {
		if items[i].File.Path < items[j].File.Path {
			return -1
		} else {
			return 1
		}
	}
	return items[i].File.Index - items[j].File.Index
}

func resultItemToString(item framework.ResultItem) string {
	return fmt.Sprintf("resource-ref:%s,field:%s,message:%s",
		item.ResourceRef, item.Field, item.Message)
}

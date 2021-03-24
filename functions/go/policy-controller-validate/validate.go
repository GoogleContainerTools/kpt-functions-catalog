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
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/open-policy-agent/frameworks/constraint/pkg/apis/templates/v1beta1"
	opaclient "github.com/open-policy-agent/frameworks/constraint/pkg/client"
	"github.com/open-policy-agent/frameworks/constraint/pkg/client/drivers/local"
	"github.com/open-policy-agent/frameworks/constraint/pkg/core/templates"
	opatypes "github.com/open-policy-agent/frameworks/constraint/pkg/types"
	"github.com/open-policy-agent/gatekeeper/pkg/target"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// SourcePathAnnotation is the path to the configuration file containing the object.
const SourcePathAnnotation = "config.kubernetes.io/path"

var scheme = runtime.NewScheme()

func init() {
	err := v1beta1.SchemeBuilder.AddToScheme(scheme)
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
		ct, isConstraintTemplate := obj.(*v1beta1.ConstraintTemplate)
		if !isConstraintTemplate {
			continue
		}
		templ := &templates.ConstraintTemplate{}
		if err := scheme.Convert(ct, templ, nil); err != nil {
			return nil, err
		}
		templs = append(templs, templ)
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
func Validate(objects []runtime.Object) error {
	client, err := createClient()
	if err != nil {
		return err
	}
	tmpls, err := gatherTemplates(objects)
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, t := range tmpls {
		if _, err = client.AddTemplate(ctx, t); err != nil {
			return err
		}
	}
	cstrs, err := gatherConstraints(objects)
	if err != nil {
		return err
	}
	for _, c := range cstrs {
		if _, err = client.AddConstraint(ctx, c); err != nil {
			return err
		}
	}

	for _, obj := range objects {
		if _, err = client.AddData(ctx, obj); err != nil {
			return err
		}
	}

	resps, err := client.Audit(ctx)
	if err != nil {
		return err
	}
	results := resps.Results()
	if len(results) > 0 {
		return parseResults(results)
	}
	return nil
}

func parseResults(results []*opatypes.Result) error {
	var msgs []string
	for _, r := range results {
		u, ok := r.Resource.(*unstructured.Unstructured)
		if !ok {
			return fmt.Errorf("could not cast to unstructured: %+v", r.Resource)
		}
		name := u.GetName()
		path, found := u.GetAnnotations()[SourcePathAnnotation]
		if !found {
			path = "?"
		}
		constraintName := r.Constraint.GetName()
		msgs = append(msgs, fmt.Sprintf("%s\n\nname: %q\npath: %s\nviolatedConstraint: %s", r.Msg, name, path, constraintName))
	}

	sort.Strings(msgs)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d violations:\n\n", len(results)))
	for i, m := range msgs {
		sb.WriteString(fmt.Sprintf("[%d] %s\n\n", i+1, m))
	}
	return errors.New(sb.String())
}

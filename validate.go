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

package gatekeeper

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/pkg/framework/constants"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/pkg/framework/io"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/pkg/framework/types"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/pkg/functions/util"
	"github.com/open-policy-agent/frameworks/constraint/pkg/apis/templates/v1beta1"
	opaclient "github.com/open-policy-agent/frameworks/constraint/pkg/client"
	"github.com/open-policy-agent/frameworks/constraint/pkg/client/drivers/local"
	"github.com/open-policy-agent/frameworks/constraint/pkg/core/templates"
	opatypes "github.com/open-policy-agent/frameworks/constraint/pkg/types"
	"github.com/open-policy-agent/gatekeeper/pkg/target"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var scheme = runtime.NewScheme()

func init() {
	io.Register(v1beta1.SchemeGroupVersion.WithKind("ConstraintTemplate"), func() types.KubernetesObject {
		return &v1beta1.ConstraintTemplate{}
	})
	err := v1beta1.AddToSchemes.AddToScheme(scheme)
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

func gatherTemplates(configs *types.Configs) ([]*templates.ConstraintTemplate, error) {
	var templs []*templates.ConstraintTemplate
	for _, obj := range *configs {
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

func gatherConstraints(configs *types.Configs) ([]types.KubernetesObject, error) {
	var cstrs []types.KubernetesObject
	for _, obj := range *configs {
		gvk := obj.GroupVersionKind()
		if gvk.Group == "constraints.gatekeeper.sh" {
			cstrs = append(cstrs, obj)
		}
	}
	return cstrs, nil
}

// Validate makes sure the configs passed to it comply with any Constraints and
// Constraint Templates present in the list of configs
func Validate(configs *types.Configs) error {
	client, err := createClient()
	if err != nil {
		return err
	}
	tmpls, err := gatherTemplates(configs)
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, t := range tmpls {
		if _, err = client.AddTemplate(ctx, t); err != nil {
			return err
		}
	}
	cstrs, err := gatherConstraints(configs)
	if err != nil {
		return err
	}
	for _, c := range cstrs {
		u, err2 := util.ToUnstructured(c)
		if err2 != nil {
			return err2
		}
		if _, err2 = client.AddConstraint(ctx, u); err2 != nil {
			return err2
		}
	}

	for _, o := range *configs {
		jsn, err2 := json.Marshal(o)
		if err2 != nil {
			return err2
		}

		u := unstructured.Unstructured{}
		err = u.UnmarshalJSON(jsn)
		if err != nil {
			return err
		}

		if _, err = client.AddData(ctx, u); err != nil {
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
		path, found := u.GetAnnotations()[constants.SourcePathAnnotation]
		if !found {
			path = "?"
		}
		msgs = append(msgs, fmt.Sprintf("%s\n\nname: %q\npath: %s", r.Msg, name, path))
	}

	sort.Strings(msgs)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d violations:\n\n", len(results)))
	for i, m := range msgs {
		sb.WriteString(fmt.Sprintf("[%d] %s\n\n", i+1, m))
	}
	return types.NewConfigError(sb.String())
}

var _ types.ConfigFunc = Validate

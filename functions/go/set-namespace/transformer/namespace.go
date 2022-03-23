// Copyright 2022 Google LLC
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
package transformer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

const (
	fnConfigAPIVersion  = "fn.kpt.dev/v1alpha1"
	fnConfigKind        = "SetNamespace"
	dependsOnAnnotation = "config.kubernetes.io/depends-on"
	groupIdx            = 0
	namespaceIdx        = 2
	kindIdx             = 3
	nameIdx             = 4
)

var (
	// <group>/namespaces/<namespace>/<kind>/<name>
	namespacedResourcePattern = regexp.MustCompile(`\A([-.\w]*)/namespaces/([-.\w]*)/([-.\w]*)/([-.\w]*)\z`)
	dependsOnKeyPattern       = func(group, kind, name string) string {
		return fmt.Sprintf("%s/%s/%s", group, kind, name)
	}
)

type Subject struct {
	Kind      string `json:"kind,omitempty" yaml:"kind,omitempty"`
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	ApiGroup  string `json:"apiGroup,omitempty" yaml:"apiGroup,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}


func SetNamespace(rl *fn.ResourceList) error {
	tc := NamespaceTransformer{}
	// Get "namespace" arguments from FunctionConfig
	err := tc.Config(rl.FunctionConfig)
	if err != nil {
		rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, rl.FunctionConfig))
		return nil
	}
	// Update "namespace" to the proper resources.
	tc.Transform(rl.Items)
	var result *fn.Result
	if len(tc.Errors) != 0 {
		errMsg := strings.Join(tc.Errors, "\n")
		result = fn.GeneralResult(errMsg, fn.Error)
	} else {
		result = fn.GeneralResult("namespace updated", fn.Info)
	}
	rl.Results = append(rl.Results, result)
	return nil
}

type NamespaceTransformer struct {
	Namespace    string
	DependsOnMap map[string]bool
	Errors       []string
}

func (p *NamespaceTransformer) Config(o *fn.KubeObject) error {
	switch {
	case o.IsGVK("v1", "ConfigMap"):
		p.Namespace = o.GetStringOrDie("data", "namespace")
		if p.Namespace == "" {
			return fmt.Errorf("`data.namespace` should not be empty")
		}
	case o.IsGVK(fnConfigAPIVersion, fnConfigKind):
		p.Namespace = o.GetStringOrDie("namespace")
		if p.Namespace == "" {
			return fmt.Errorf("`namespace` should not be empty")
		}
	case o.IsGVK("v1", "ConfigMap") && o.GetName() == "kptfile.kpt.dev":
		p.Namespace = o.GetStringOrDie("data", "name")
		if p.Namespace == "" {
			return fmt.Errorf("`data.name` should not be empty")
		}
	default:
		return fmt.Errorf("unknown functionConfig Kind=%v ApiVersion=%v, expect `%v` or `ConfigMap`",
			o.GetKind(), o.GetAPIVersion(), fnConfigKind)
	}
	return nil
}

func (p *NamespaceTransformer) Transform(objects []*fn.KubeObject) {
	p.SetDependsOnMap(objects)
	for _, o := range objects {
		switch {
		case o.IsGVK("v1", "Namespace"):
			o.SetName(p.Namespace)
		case o.IsGVK("apiextensions.k8s.io/v1", "CustomResourceDefinition"):
			err := o.Set(p.Namespace, "spec", "conversion", "webhook", "clientConfig", "service", "namespace")
			if err != nil {
				p.Errors = append(p.Errors, err.Error())
			}
		case o.IsGVK("apiregistration.k8s.io/v1", "APIService"):
			err := o.Set(p.Namespace, "spec", "service", "namespace")
			if err != nil {
				p.Errors = append(p.Errors, err.Error())
			}
		case o.GetKind() == "ClusterRoleBinding" || o.GetKind() == "RoleBinding":
			var subjects []*Subject
			o.GetOrDie(&subjects, "subjects")
			for _, s := range subjects {
				if s.Name == "default" {
					s.Namespace = p.Namespace
				}
			}
			o.SetOrDie(&subjects, "subjects")
		case o.HasNamespace():
			// Only update the namespace scoped resource. To determine if a resource is namespace scoped,
			// we assume its namespace should have been set.
			o.SetNamespace(p.Namespace)
		default:
			// skip the cluster scoped resource. We determine if a resource is cluster scoped by
			// checking if the metadata.namespace is configured.
		}
		// Shall we accept custom fieldspec?
		// p.SetCustomFieldSpec(o)
		p.UpdateAnnotation(o)
	}
}

func (p *NamespaceTransformer) SetDependsOnMap(objects []*fn.KubeObject) {
	p.DependsOnMap = map[string]bool{}
	for _, o := range objects {
		group := o.GetAPIVersion()
		if i := strings.Index(o.GetAPIVersion(), "/"); i > -1 {
			group = group[:i]
		}
		key := dependsOnKeyPattern(group, o.GetKind(), o.GetName())
		p.DependsOnMap[key] = true
	}
}

func (p *NamespaceTransformer) UpdateAnnotation(o *fn.KubeObject) {
	anno, ok := o.GetAnnotations()[dependsOnAnnotation]
	if !ok {
		return
	}
	if !namespacedResourcePattern.MatchString(anno) {
		return
	}
	segments := strings.Split(anno, "/")
	dependsOnkey := dependsOnKeyPattern(segments[groupIdx], segments[kindIdx], segments[nameIdx])
	if ok := p.DependsOnMap[dependsOnkey]; ok {
		segments[namespaceIdx] = p.Namespace
		newAnno := strings.Join(segments, "/")
		o.SetAnnotation(dependsOnAnnotation, newAnno)
	}
}

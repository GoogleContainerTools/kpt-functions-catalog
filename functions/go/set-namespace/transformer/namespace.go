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
	"reflect"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

func SetNamespace(rl *fn.ResourceList) error {
	tc := NamespaceTransformer{}
	// Get "namespace" arguments from FunctionConfig
	err := tc.Config(rl.FunctionConfig)
	if err != nil {
		rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, rl.FunctionConfig))
		return nil
	}
	if tc.validate(rl.Items) {
		// Update "namespace" to the proper resources.
		tc.Transform(rl.Items)
	}
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

// Config gets the attributes from different FunctionConfig formats.
func (p *NamespaceTransformer) Config(o *fn.KubeObject) error {
	switch {
	case o.IsGVK("v1", "ConfigMap"):
		p.NewNamespace = o.GetStringOrDie("data", "namespace")
		if p.NewNamespace == "" {
			return fmt.Errorf("`data.namespace` should not be empty")
		}
		p.MatchingNamespace = o.GetStringOrDie("data", "namespaceSelector")
	case o.IsGVK(fnConfigAPIVersion, fnConfigKind):
		p.NewNamespace = o.GetStringOrDie("namespace")
		if p.NewNamespace == "" {
			return fmt.Errorf("`namespace` should not be empty")
		}
		p.MatchingNamespace = o.GetStringOrDie("data", "namespaceSelector")
	case o.IsGVK("v1", "ConfigMap") && o.GetName() == "kptfile.kpt.dev":
		p.NewNamespace = o.GetStringOrDie("data", "name")
		if p.NewNamespace == "" {
			return fmt.Errorf("`data.name` should not be empty")
		}
	default:
		return fmt.Errorf("unknown functionConfig Kind=%v ApiVersion=%v, expect `%v` or `ConfigMap`",
			o.GetKind(), o.GetAPIVersion(), fnConfigKind)
	}
	return nil
}

// Transform replace the existing Namespace object, namespace-scoped resources' `metadata.name` and
// other namespace reference fields to the new value.
func (p *NamespaceTransformer) Transform(objects []*fn.KubeObject) {
	p.SetDependsOnMap(objects)
	for _, o := range objects {
		switch {
		// Skip local resource which `kpt live apply` skips.
		case o.IsLocalConfig():
			continue
		case o.IsGVK("v1", "Namespace"):
			o.SetName(p.NewNamespace)
		case o.IsGVK("apiextensions.k8s.io/v1", "CustomResourceDefinition"):
			p.updateNamespace(o, "spec", "conversion", "webhook", "clientConfig", "service", "namespace")
		case o.IsGVK("apiregistration.k8s.io/v1", "APIService"):
			p.updateNamespace(o, "spec", "service", "namespace")
		case o.GetKind() == "ClusterRoleBinding" || o.GetKind() == "RoleBinding":
			var subjects []*Subject
			o.GetOrDie(&subjects, "subjects")
			for _, s := range subjects {
				if s.Kind == "ServiceAccount" {
					if p.isSelectorMode() && s.Namespace != p.MatchingNamespace {
						continue
					}
					s.Namespace = p.NewNamespace
				}
			}
			o.SetOrDie(&subjects, "subjects")
		case o.HasNamespace():
			// Only update the namespace scoped resource. To determine if a resource is namespace scoped,
			// we assume its namespace should have been set.
			p.updateNamespace(o, "metadata", "namespace")
		default:
			// skip the cluster scoped resource. We determine if a resource is cluster scoped by
			// checking if the metadata.namespace is configured.
		}
		p.UpdateAnnotation(o)
	}
}

// updateNamespace updates the namespace in three modes:
// - Restrict mode: always update namespace from the given `fieldpaths`
// - NsObjectSelector mode: update the namespace which matches the Namespace object found in resourceList.items.
// - CustomSelector mode: update the namespace which matches the `namespaceSelector` in resourceList.functionConfig.
func (p *NamespaceTransformer) updateNamespace(o *fn.KubeObject, fieldPaths ...string) {
	if p.isSelectorMode() {
		ns := o.GetStringOrDie(fieldPaths...)
		if ns != p.MatchingNamespace {
			return
		}
	}
	if err := o.Set(p.NewNamespace, fieldPaths...); err != nil {
		p.Errors = append(p.Errors, err.Error())
	}
}

// isSelectorMode decides whether the namespace fields should match the given namespace
// before updating.
func (p *NamespaceTransformer) isSelectorMode() bool {
	switch p.Mode {
	case CustomSelector:
		return true
	case NsObjectSelector:
		return true
	default: //  Restrict Mode
		return false
	}
}

// validate guarantees the input resourceList.items satisfy the UpdateMode requirements.
func (p *NamespaceTransformer) validate(objects []*fn.KubeObject) bool {
	nsCount, fromResources := p.listCurrentNamespaces(objects)
	// Check "CustomSelector" Mode
	if p.MatchingNamespace != "" {
		p.Mode = CustomSelector
		if len(nsCount) == 0 {
			return true
		}
		p.Errors = append(p.Errors,
			"found Namespace objects from the input resources, "+
				"you cannot use `namespaceSelector` in FunctionConfig together with Namespace objects")
		return false
	}
	// Check "NsObjectSelector" Mode
	if len(nsCount) > 1 {
		msg := fmt.Sprintf("cannot accept more than one Namespace objects from the input resources, found %v",
			nsCount)
		p.Errors = append(p.Errors, msg)
		return false
	}
	if len(nsCount) == 1 {
		matchingNs := reflect.ValueOf(nsCount).MapKeys()[0].String()
		if nsCount[matchingNs] > 1 {
			msg := fmt.Sprintf("found more than one Namespace objects of the same name %v", matchingNs)
			p.Errors = append(p.Errors, msg)
			return false
		}
		p.Mode = NsObjectSelector
		p.MatchingNamespace = matchingNs
		return true
	}
	// Check "Restrict" Mode
	if len(fromResources) > 1 {
		msg := fmt.Sprintf("all input namespace-scoped resources should be under the same namespace "+
			"but found different namespaces: %v ", reflect.ValueOf(fromResources).MapKeys())
		p.Errors = append(p.Errors, msg)
		return false
	}
	p.Mode = Restrict
	p.MatchingNamespace = reflect.ValueOf(fromResources).MapKeys()[0].String()
	return true
}

// listCurrentNamespaces iterates the input resourcelist.items and list all namespaces found.
func (p *NamespaceTransformer) listCurrentNamespaces(objects []*fn.KubeObject) (map[string]int, map[string]bool) {
	nsCount := map[string]int{}
	fromResources := map[string]bool{}
	for _, o := range objects {
		// Skip local resource which `kpt live apply` skips.
		if o.IsLocalConfig() {
			continue
		}
		switch {
		case o.IsGVK("v1", "Namespace"):
			nsCount[o.GetName()] += 1
			fromResources[o.GetName()] = true
		case o.IsGVK("apiextensions.k8s.io/v1", "CustomResourceDefinition"):
			ns := o.GetStringOrDie("spec", "conversion", "webhook", "clientConfig", "service", "namespace")
			fromResources[ns] = true
		case o.IsGVK("apiregistration.k8s.io/v1", "APIService"):
			ns := o.GetStringOrDie("spec", "service", "namespace")
			fromResources[ns] = true
		case o.GetKind() == "ClusterRoleBinding" || o.GetKind() == "RoleBinding":
			var subjects []*Subject
			o.GetOrDie(&subjects, "subjects")
			for _, s := range subjects {
				if s.Kind == "ServiceAccount" {
					fromResources[s.Namespace] = true
				}
			}
		case o.HasNamespace():
			fromResources[o.GetNamespace()] = true
		}
	}
	return nsCount, fromResources
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
		if p.isSelectorMode() && segments[namespaceIdx] != p.MatchingNamespace {
			return
		}
		segments[namespaceIdx] = p.NewNamespace
		newAnno := strings.Join(segments, "/")
		o.SetAnnotation(dependsOnAnnotation, newAnno)
	}
}

type Subject struct {
	Kind      string `json:"kind,omitempty" yaml:"kind,omitempty"`
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	ApiGroup  string `json:"apiGroup,omitempty" yaml:"apiGroup,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

type NamespaceTransformer struct {
	NewNamespace      string
	MatchingNamespace string
	Mode              UpdateMode
	DependsOnMap      map[string]bool
	Errors            []string
}

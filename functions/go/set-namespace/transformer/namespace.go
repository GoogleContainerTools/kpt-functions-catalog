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
	"k8s.io/apimachinery/pkg/util/sets"
)

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

// Config gets the attributes from different FunctionConfig formats.
func (p *NamespaceTransformer) Config(o *fn.KubeObject) error {
	switch {
	case o.IsGVK("v1", "ConfigMap"):
		p.NewNamespace = o.GetStringOrDie("data", "namespace")
		if p.NewNamespace == "" {
			return fmt.Errorf("`data.namespace` should not be empty")
		}
		p.namespaceSelector = o.GetStringOrDie("data", "namespaceSelector")
	case o.IsGVK(fnConfigAPIVersion, fnConfigKind):
		p.NewNamespace = o.GetStringOrDie("namespace")
		if p.NewNamespace == "" {
			return fmt.Errorf("`namespace` should not be empty")
		}
		p.namespaceSelector = o.GetStringOrDie("data", "namespaceSelector")
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
	namespaces, nsObjCounter := FindAllNamespaces(objects)
	if oldNs, ok := p.GetOldNamespace(namespaces, nsObjCounter); ok {
		ReplaceNamespace(objects, oldNs, p.NewNamespace)
		// Update the resources annotation "config.kubernetes.io/depends-on" which may contain old namespace value.
		dependsOnMap := GetDependsOnMap(objects)
		UpdateAnnotation(objects, oldNs, p.NewNamespace, dependsOnMap)
	}
}

// VisitNamespaces iterates the `objects` to execute the `callback` function on each corresponding namespace field.
func VisitNamespaces(objects []*fn.KubeObject, callback func(ns *Namespace)) {
	for _, o := range objects {
		switch {
		// Skip local resource which `kpt live apply` skips.
		case o.IsLocalConfig():
			continue
		case o.IsGVK("v1", "Namespace"):
			ns := NewNamespace(o, []string{"metadata", "name"})
			callback(ns)
		case o.IsGVK("apiextensions.k8s.io/v1", "CustomResourceDefinition"):
			fields := []string{"spec", "conversion", "webhook", "clientConfig", "service", "namespace"}
			callback(NewNamespace(o, fields))
		case o.IsGVK("apiregistration.k8s.io/v1", "APIService"):
			fields := []string{"spec", "service", "namespace"}
			callback(NewNamespace(o, fields))
		/* TODO(yuwenma): waiting for the new features in kpt-functions-sdk to merge in.
		case o.GetKind() == "ClusterRoleBinding" || o.GetKind() == "RoleBinding":
		subjects := o.GetSliceOrDie("subjects")
		for _, s := range subjects {
			if s.GetKind() == "ServiceAccount" {
				ns := s.GetNamespace()
				visitor(ns, false)
				nsCollector <- func() {
					nsPtr := &ns
					p.updateNamespace(nsPtr)
					s.SetNamespace(*nsPtr)
					o.SetOrDie(&subjects, "subjects")
				}
			}
		}
		*/
		case o.HasNamespace():
			// Only update the namespace scoped resource. To determine if a resource is namespace scoped,
			// we assume its namespace should have been set.
			fields := []string{"metadata", "namespace"}
			callback(NewNamespace(o, fields))
		default:
			// skip the cluster scoped resource. We determine if a resource is cluster scoped by
			// checking if the metadata.namespace is configured.
		}
	}
}

// FindAllNamespaces iterates the `objects` to list all namespaces and count the number of Namespace objects.
func FindAllNamespaces(objects []*fn.KubeObject) ([]string, map[string]int) {
	nsObjCounter := map[string]int{}
	namespaces := sets.NewString()
	VisitNamespaces(objects, func(ns *Namespace) {
		if ns.Name == "" {
			return
		}
		if ns.IsNamespace {
			nsObjCounter[ns.Name] += 1
		}
		namespaces.Insert(ns.Name)
	})
	return namespaces.List(), nsObjCounter
}

// ReplaceNamespace iterates the `objects` to replace the `OldNs` with `newNs` on namespace field.
func ReplaceNamespace(objects []*fn.KubeObject, oldNs, newNs string) {
	VisitNamespaces(objects, func(ns *Namespace) {
		if ns.Name == "" {
			return
		}
		if ns.Name == oldNs {
			ns.Set(newNs)
		}
	})
}

// GetDependsOnMap iterates `objects` to get the annotation which contains namespace value.
func GetDependsOnMap(objects []*fn.KubeObject) map[string]bool {
	dependsOnMap := map[string]bool{}
	VisitNamespaces(objects, func(ns *Namespace) {
		key := ns.GetDependsOnAnno()
		dependsOnMap[key] = true
	})
	return dependsOnMap
}

// UpdateAnnotation updates the `objects`'s "config.kubernetes.io/depends-on" annotation which contains namespace value.
func UpdateAnnotation(objects []*fn.KubeObject, oldNs, newNs string, dependsOnMap map[string]bool) {
	VisitNamespaces(objects, func(ns *Namespace) {
		if ns.DependsOnAnno == "" || !namespacedResourcePattern.MatchString(ns.DependsOnAnno) {
			return
		}
		segments := strings.Split(ns.DependsOnAnno, "/")
		dependsOnkey := dependsOnKeyPattern(segments[groupIdx], segments[kindIdx], segments[nameIdx])
		if ok := dependsOnMap[dependsOnkey]; ok {
			if segments[namespaceIdx] == oldNs {
				segments[namespaceIdx] = newNs
				newAnno := strings.Join(segments, "/")
				ns.SetDependsOnAnno(newAnno)
			}
		}
	})
}

// GetOldNamespace finds the existing namespace and make sure the input resourceList.items satisfy the requirements.
// - no more than one Namespace Object can exist in the input resource.items
// - If Namespace object exists, its name is the `oldNs`
// - If `namespaceSelector` is given, its value is the `oldNs`
// - If neither Namespace object nor `namespaceSelector` found, all resources should have the same namespace value and
// this value is teh `oldNs`
func (p *NamespaceTransformer) GetOldNamespace(fromResources []string, nsCount map[string]int) (string, bool) {
	if p.namespaceSelector != "" {
		if len(nsCount) == 0 {
			return p.namespaceSelector, true
		}
		p.Errors = append(p.Errors,
			"found Namespace objects from the input resources, "+
				"you cannot use `namespaceSelector` in FunctionConfig together with Namespace objects")
		return "", false
	}
	if len(nsCount) > 1 {
		msg := fmt.Sprintf("cannot accept more than one Namespace objects from the input resources, found %v",
			nsCount)
		p.Errors = append(p.Errors, msg)
		return "", false
	}
	if len(nsCount) == 1 {
		// Use the namespace object as the matching namespace if `namespaceSelector` is not given.
		oldNs := reflect.ValueOf(nsCount).MapKeys()[0].String()
		if nsCount[oldNs] > 1 {
			msg := fmt.Sprintf("found more than one Namespace objects of the same name %v", oldNs)
			p.Errors = append(p.Errors, msg)
			return "", false
		}
		return oldNs, true
	}
	if len(fromResources) > 1 {
		msg := fmt.Sprintf("all input namespace-scoped resources should be under the same namespace "+
			"but found different namespaces: %v ", strings.Join(fromResources, ","))
		p.Errors = append(p.Errors, msg)
		return "", false
	}
	return fromResources[0], true
}

type NamespaceTransformer struct {
	NewNamespace      string
	namespaceSelector string
	DependsOnMap      map[string]bool
	Errors            []string
}

func NewNamespace(obj *fn.KubeObject, path []string) *Namespace {
	setter := func(newNs string) {
		obj.Set(newNs, path...)
	}
	annoSetter := func(newAnno string) {
		obj.SetAnnotation(dependsOnAnnotation, newAnno)
	}
	annoGetter := func() string {
		group := obj.GetAPIVersion()
		if i := strings.Index(obj.GetAPIVersion(), "/"); i > -1 {
			group = group[:i]
		}
		return dependsOnKeyPattern(group, obj.GetKind(), obj.GetName())
	}
	return &Namespace{
		Name:          obj.GetStringOrDie(path...),
		IsNamespace:   obj.IsGVK("v1", "Namespace"),
		DependsOnAnno: obj.GetAnnotations()[dependsOnAnnotation],
		setter:        setter,
		annoGetter:    annoGetter,
		annoSetter:    annoSetter,
	}
}

type Namespace struct {
	Name          string
	IsNamespace   bool
	DependsOnAnno string
	obj           *fn.KubeObject
	setter        func(newNs string)
	annoGetter    func() string
	annoSetter    func(newDependsOnAnno string)
}

func (n *Namespace) Set(newNs string) {
	n.setter(newNs)
}

func (n *Namespace) SetDependsOnAnno(newDependsOnAnno string) {
	n.annoSetter(newDependsOnAnno)
}

func (n *Namespace) GetDependsOnAnno() string {
	return n.annoGetter()
}

// Copyright 2021 Google LLC
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

package terraformgenerator

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	sdk "github.com/GoogleContainerTools/kpt-functions-catalog/thirdparty/kyaml/fnsdk"
)

type terraformResources struct {
	resources map[string]*terraformResource
	grouped   map[string][]*terraformResource
	Variables map[string]*variable
}

func (rs *terraformResources) GetVersion() string {
	return "0.1.0"
}

func (rs *terraformResources) getResourceRef(kind string, name string, item *sdk.KubeObject) (*terraformResource, error) {
	if rs.resources == nil {
		rs.resources = map[string]*terraformResource{}
	}

	key := fmt.Sprintf("%s/%s", kind, name)
	resourceRef, found := rs.resources[key]
	if !found {
		resourceRef = &terraformResource{
			Name:      name,
			Kind:      kind,
			resources: rs,
		}
		rs.resources[key] = resourceRef
	}
	if item != nil {
		resourceRef.Item = item

		// attach parents
		err := resourceRef.attachReferences()
		if err != nil {
			return nil, err
		}
	}
	return resourceRef, nil
}

func (rs *terraformResources) getGrouped() map[string][]*terraformResource {
	if rs.grouped != nil {
		return rs.grouped
	}
	// iterate over resources in a stable order
	keys := make([]string, len(rs.resources))
	i := 0
	for k := range rs.resources {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	rs.grouped = make(map[string][]*terraformResource)

	for _, key := range keys {
		resource := rs.resources[key]
		rs.grouped[resource.Kind] = append(rs.grouped[resource.Kind], resource)
	}

	return rs.grouped
}

type terraformResource struct {
	Name       string                        // The name of the resource (from metadata.name)
	Kind       string                        // The Kubernetes Kind of the resource
	Item       *sdk.KubeObject               // The Kubernetes object for the resource, if it exists
	Parent     *terraformResource            // The parent for the resource (ex. folder for project), if there is one
	Children   []*terraformResource          // A list of any children underneath the resource
	isChild    bool                          // Whether the resource has a valid hierarchical Parent or not
	resources  *terraformResources           // A back-reference to the bundle of resources this resource is part of
	variable   *variable                     // If this resource is defined by a variable, a reference to the associated variable
	References map[string]*terraformResource // A map of resources this resource references, by the kind of reference
}

// Return if the resource itself should be created
func (resource *terraformResource) ShouldCreate() bool {
	return resource.Item != nil
}

// Retrieve a boolean from the resource
func (resource *terraformResource) GetBool(path ...string) bool {
	value := resource.GetStringFromObject(path...)
	if len(value) == 0 {
		return false
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		// todo: log the failure
		return false
	}
	return boolValue
}

// Look up a referenced resource at a given path
func (resource *terraformResource) GetStringFromObject(path ...string) string {
	var ref string
	found, err := resource.Item.Get(&ref, path...)
	if err != nil || !found {
		return ""
	}
	return ref
}

// Return if the resource itself should be created
func (resource *terraformResource) GetOrganization() *terraformResource {
	if resource.Parent.Kind == "Organization" {
		return resource.Parent
	}
	orgs := resource.resources.getGrouped()["Organization"]
	if len(orgs) < 1 {
		sdk.Logf("Failed to fetch organization for %s/%s", resource.Kind, resource.Name)
		return nil
	}

	return orgs[0]
}

// Attach parents and other references to a resource
func (resource *terraformResource) attachReferences() error {
	resource.References = make(map[string]*terraformResource)
	paths := [][]string{
		{"BillingAccount", "spec", "billingAccountRef", "external"},
	}
	for _, path := range paths {
		kind := path[0]
		ref := resource.getReferencedResource(kind, path[1:]...)
		if ref != nil {
			resource.References[kind] = ref
		}
	}

	// attach parents
	parentKind, parentName, err := resource.getParentRef()
	if err != nil {
		sdk.Logf("no parent reference found for %s, %v", resource.Item.Name(), err)
		return err
	}

	if parentKind != "" {
		parentRef, err := resource.resources.getResourceRef(parentKind, parentName, nil)
		if err != nil {
			return err
		}
		parentRef.Children = append(parentRef.Children, resource)
		resource.isChild = true
		resource.Parent = parentRef
	}

	return nil
}

// Retrieve a referenced resource from the object spec
func (resource *terraformResource) getReferencedResource(kind string, path ...string) *terraformResource {
	name := strings.TrimSpace(resource.GetStringFromObject(path...))
	if len(name) == 0 {
		return nil
	}

	ref, err := resource.resources.getResourceRef(kind, name, nil)
	if err != nil {
		sdk.Logf("Failed to retrieve resource reference for %s/%s from %s", kind, name, resource.Name)
	}
	ref.Children = append(ref.Children, resource)

	return ref
}

func (resource *terraformResource) getParentRef(path ...string) (string, string, error) {
	paths := [][]string{
		{"Folder", "spec", "folderRef", "name"},
		{"Folder", "spec", "folderRef", "external"},
		{"Organization", "spec", "organizationRef", "external"},
		{"Project", "metadata", "annotations", "cnrm.cloud.google.com/my-project"},
		{"detect", "spec", "resourceRef", "external"},
		{"detect", "spec", "resourceRef", "name"},
	}

	for _, path := range paths {
		name := resource.GetStringFromObject(path[1:]...)
		if name == "" {
			continue
		}

		kind := path[0]
		if kind == "detect" {
			kind = resource.GetStringFromObject(append(path[1:len(path)-1], "kind")...)
		}

		return kind, strings.TrimSpace(name), nil
	}

	return "", "", nil
}

func (ref *terraformResource) GetDisplayName() string {
	var displayName string
	found, err := ref.Item.Get(&displayName, "spec", "displayName")
	if err == nil && found {
		return displayName
	}
	found, err = ref.Item.Get(&displayName, "spec", "name")
	if err == nil && found {
		return displayName
	}
	// TODO: log failure to find
	return ref.Item.Name()
}

var tfNameRegex = regexp.MustCompile(`[^a-zA-Z\d_-]`)

func (ref *terraformResource) GetResourceName() string {
	// For real resources, use their name
	if ref.ShouldCreate() {
		name := ref.Name
		if name != "" {
			name = tfNameRegex.ReplaceAllString(name, "-")
			return name
		}
	}

	ofKind := ref.resources.getGrouped()[ref.Kind]
	if len(ofKind) < 2 {
		return strings.ToLower(ref.Kind)
	}
	for i, testResource := range ofKind {
		if testResource.Name == ref.Name {
			return fmt.Sprintf("%s-%d", strings.ToLower(ref.Kind), i+1)
		}
	}
	return ""
}

func (ref *terraformResource) GetTerraformId(prefix ...bool) string {
	if ref.ShouldCreate() {
		return fmt.Sprintf("google_folder.%s.name", ref.GetResourceName())
	}
	hasVariable := ref.variable != nil
	usePrefix := !(len(prefix) > 0 && !prefix[0])
	isOrg := ref.Kind == "Organization"

	switch {
	case !usePrefix && hasVariable:
		return fmt.Sprintf("var.%s", ref.variable.Name)
	case !usePrefix && !hasVariable:
		return fmt.Sprintf("\"%s\"", ref.Name)
	case isOrg && hasVariable:
		return fmt.Sprintf("\"organizations/${var.%s}\"", ref.variable.Name)
	case isOrg && !hasVariable:
		return fmt.Sprintf("\"organizations/%s\"", ref.Name)
	case hasVariable:
		return fmt.Sprintf("\"folders/${var.%s}\"", ref.variable.Name)
	default:
		return fmt.Sprintf(`"folders/%s"`, ref.Name)
	}
}

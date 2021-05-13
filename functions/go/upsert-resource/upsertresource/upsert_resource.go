package upsertresource

import (
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// UpsertResource upserts resource to input list of resources
// resources are uniquely identifies by Group, Kind, Name and Namespace
type UpsertResource struct {
	// Resource is the input resource for upsert
	Resource *yaml.RNode
}

// Filter implements UpsertResource as a yaml.Filter
func (ur UpsertResource) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	replaced, err := ReplaceResource(nodes, ur.Resource)
	if err != nil {
		return nodes, err
	}
	if !replaced {
		return AddResource(nodes, ur.Resource)
	}
	return nodes, nil
}

// ReplaceResource checks if the inputResource matches with any of the resources in nodes list
// and replaces it with inputResource
// nodes are matched based on Group, Kind, Name and Namespace
func ReplaceResource(nodes []*yaml.RNode, inputResource *yaml.RNode) (bool, error) {
	found := false
	inputMeta, err := inputResource.GetMeta()
	if err != nil {
		return false, err
	}
	for i := range nodes {
		rMeta, err := nodes[i].GetMeta()
		if err != nil {
			return false, err
		}
		// skip processing resource if it is a local config
		if IsLocalConfig(rMeta) {
			continue
		}
		// check if there is a match and replace the resource
		if IsSameResource(inputMeta, rMeta) {
			nodes[i], err = deepCopy(inputResource)
			if err != nil {
				return false, err
			}
			a := combineInputAndMatchedAnnotations(inputMeta.Annotations, rMeta.Annotations)
			err = nodes[i].SetAnnotations(a)
			if err != nil {
				return false, err
			}
			// found a matching resource
			found = true
		}
	}
	return found, nil
}

// AddResource appends the inputResource to the list of input nodes
// it also cleans up the meta annotations so that resource is created in new file
// by the function orchestrator
func AddResource(nodes []*yaml.RNode, inputResource *yaml.RNode) ([]*yaml.RNode, error) {
	new, err := deepCopy(inputResource)
	if err != nil {
		return nodes, err
	}
	meta, err := new.GetMeta()
	if err != nil {
		return nodes, err
	}
	// remove local, function, path and index annotations from the result
	// removing path and index annotations makes orchestrator write resource
	// to a new file
	cleanedAnno := removeMetaAnnotations(meta.Annotations)
	err = new.SetAnnotations(cleanedAnno)
	if err != nil {
		return nodes, err
	}
	nodes = append(nodes, new)
	return nodes, nil
}

// IsSameResource returns true if metadata of two resources
// have same Group, Kind, Name, Namespace
func IsSameResource(meta1, meta2 yaml.ResourceMeta) bool {
	g1, _ := ParseGroupVersion(meta1.APIVersion)
	g2, _ := ParseGroupVersion(meta1.APIVersion)
	return g1 == g2 && meta1.Kind == meta2.Kind &&
		meta1.Name == meta2.Name &&
		meta1.Namespace == meta2.Namespace
}

// ParseGroupVersion parses a KRM metadata apiVersion field.
func ParseGroupVersion(apiVersion string) (group, version string) {
	if i := strings.Index(apiVersion, "/"); i > -1 {
		return apiVersion[:i], apiVersion[i+1:]
	}
	return "", apiVersion
}

// combineInputAndMatchedAnnotations combines user provided non-meta annotations from inputResource,
// with path and index annotations from matched resource and returns the result
func combineInputAndMatchedAnnotations(inputResourceAnno, matchedResourceAnno map[string]string) map[string]string {
	if inputResourceAnno == nil {
		inputResourceAnno = make(map[string]string)
	}
	if matchedResourceAnno == nil {
		matchedResourceAnno = make(map[string]string)
	}
	res := make(map[string]string)
	// retain the annotations from the input resource in fn-config,
	// these should be written to matched resource
	for k, v := range inputResourceAnno {
		res[k] = v
	}
	// retain the path and index annotation from matched resource to result
	res[kioutil.PathAnnotation] = matchedResourceAnno[kioutil.PathAnnotation]
	res[kioutil.IndexAnnotation] = matchedResourceAnno[kioutil.IndexAnnotation]
	// remove local and function meta annotations from the result as they should
	// not be written to output resource
	return removeLocalAndFnAnnotations(res)
}

// deepCopy returns the deep copy of the input RNode
func deepCopy(node *yaml.RNode) (*yaml.RNode, error) {
	// serialize input RNode to string
	s, err := node.String()
	if err != nil {
		return nil, err
	}
	// create new RNode from yaml string
	res, err := yaml.Parse(s)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// removeMetaAnnotations cleans index, path, local and fn annotations
func removeMetaAnnotations(a map[string]string) map[string]string {
	a = removeLocalAndFnAnnotations(a)
	delete(a, kioutil.PathAnnotation)
	delete(a, kioutil.IndexAnnotation)
	return a
}

// removeLocalAndFnAnnotations cleans local and fn annotations
func removeLocalAndFnAnnotations(a map[string]string) map[string]string {
	delete(a, filters.LocalConfigAnnotation)
	delete(a, runtimeutil.FunctionAnnotationKey)
	// using hard coded key as this annotation is deprecated and not exposed by kyaml
	delete(a, "config.k8s.io/function")
	return a
}

// IsLocalConfig returns true if input resource meta has local config annotation set to true
func IsLocalConfig(rMeta yaml.ResourceMeta) bool {
	return rMeta.Annotations != nil && rMeta.Annotations[filters.LocalConfigAnnotation] == "true"
}

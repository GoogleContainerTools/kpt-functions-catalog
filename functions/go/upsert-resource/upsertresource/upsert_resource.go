package upsertresource

import (
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
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
		// skip processing resource if it is a function config
		// TODO: remove this check after we stop support for v0.X
		if IsFunctionConfig(rMeta) {
			continue
		}
		// check if there is a match and replace the resource
		if IsSameResource(inputMeta, rMeta) {
			nodes[i] = inputResource.Copy()
			a := combineInputAndMatchedAnnotations(inputMeta.Annotations, rMeta.Annotations)
			err = nodes[i].SetAnnotations(a)
			if err != nil {
				return false, err
			}
			// found a matching resource
			// but continue to replace other instances of the resource
			found = true
		}
	}
	return found, nil
}

// AddResource appends the inputResource to the list of input nodes
// it also cleans up the meta annotations so that resource is created in new file
// by the function orchestrator
func AddResource(nodes []*yaml.RNode, inputResource *yaml.RNode) ([]*yaml.RNode, error) {
	new := inputResource.Copy()
	meta, err := new.GetMeta()
	if err != nil {
		return nodes, err
	}
	// remove function, path and index annotations from the result
	// removing path and index annotations makes orchestrator write resource
	// to a new file
	removeFnPathIndexAnnotations(meta.Annotations)
	err = new.SetAnnotations(meta.Annotations)
	if err != nil {
		return nodes, err
	}
	nodes = append(nodes, new)
	return nodes, nil
}

// IsSameResource returns true if metadata of two resources
// have same Group, Kind, Name, Namespace
// TODO: phanimarupaka move this to common util https://github.com/GoogleContainerTools/kpt/issues/2043
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
	// remove function meta annotations from the result as they should
	// not be written to output resource
	removeFnAnnotations(res)
	return res
}

// removeFnPathIndexAnnotations removes index, path and fn annotations
func removeFnPathIndexAnnotations(a map[string]string) {
	removeFnAnnotations(a)
	delete(a, kioutil.PathAnnotation)
	delete(a, kioutil.IndexAnnotation)
}

// removeFnAnnotations removes fn annotations
// TODO: phanimarupaka remove this method after we drop support for kpt 0.X
func removeFnAnnotations(a map[string]string) {
	delete(a, runtimeutil.FunctionAnnotationKey)
	// using hard coded key as this annotation is deprecated and not exposed by kyaml
	delete(a, "config.k8s.io/function")
}

// IsFunctionConfig returns true if input resource meta has function config annotation
func IsFunctionConfig(rMeta yaml.ResourceMeta) bool {
	return rMeta.Annotations != nil &&
		(rMeta.Annotations[runtimeutil.FunctionAnnotationKey] != "" || rMeta.Annotations["config.k8s.io/function"] != "")
}

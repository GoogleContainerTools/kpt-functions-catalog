package upsertresource

import (
	"strings"

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
	inputMeta, err := ur.Resource.GetMeta()
	if err != nil {
		return nodes, err
	}
	found := false
	for i := range nodes {
		rMeta, err := nodes[i].GetMeta()
		if err != nil {
			return nodes, err
		}
		// check if there is a match and replace the resource
		if IsSameResource(inputMeta, rMeta) {
			nodes[i], err = deepCopy(ur.Resource)
			if err != nil {
				return nodes, err
			}
			a := copyPathIndexAnnotations(inputMeta.Annotations, rMeta.Annotations)
			err = nodes[i].SetAnnotations(a)
			if err != nil {
				return nodes, err
			}
			// found a matching resource
			found = true
		}
	}
	if !found {
		// add resource if there is no matching resource
		nodes = append(nodes, ur.Resource)
	}
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

// copyPathIndexAnnotations copies the path and index annotations from matched resource to input resource annotations
func copyPathIndexAnnotations(inputResourceAnno, matchedResourceAnno map[string]string) map[string]string {
	if inputResourceAnno == nil {
		return make(map[string]string)
	}
	if matchedResourceAnno == nil {
		return inputResourceAnno
	}
	// copy path and index annotation from matched resource
	inputResourceAnno[kioutil.PathAnnotation] = matchedResourceAnno[kioutil.PathAnnotation]
	inputResourceAnno[kioutil.IndexAnnotation] = matchedResourceAnno[kioutil.IndexAnnotation]
	return inputResourceAnno
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

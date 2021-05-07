package main

import (
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func ProcessLocalConfigResources(resources []*kyaml.RNode, fltr kio.Filter) ([]*kyaml.RNode, error) {
	var lcResources, filteredResources []*kyaml.RNode
	for i := range resources {
		annotations, err := resources[i].GetAnnotations()
		if err != nil {
			return nil, err
		}
		if _, foundLocalConfig := annotations[konfig.IgnoredByKustomizeAnnotation]; foundLocalConfig {
			lcResources = append(lcResources, resources[i])
		} else {
			filteredResources = append(filteredResources, resources[i])
		}
	}

	out, err := fltr.Filter(filteredResources)

	return append(out, lcResources...), err
}

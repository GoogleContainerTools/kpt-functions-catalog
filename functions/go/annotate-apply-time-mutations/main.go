package main

import (
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/annotate-apply-time-mutations/annotateapplytimemutations"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	annotationKey = "config.kubernetes.io/apply-time-mutation"
)

type AnnotateApplyTimeMutationsProcessor struct{}

// processResource will recurse over all fields in a resource node to scan for comments to parse.
// This expects a reference to a k8s object node, and a filepath to that resource config
func processResource(object *yaml.RNode, resourcePath string) ([]framework.ResultItem, error) {
	var results []framework.ResultItem
	if object.IsNil() {
		return results, nil
	}

	// one fw instance per resource
	fw := annotateapplytimemutations.FieldWalker{FileName: resourcePath}
	err := object.VisitFields(func(node *yaml.MapNode) error {
		return fw.VisitFields(node.Value, node.Key.YNode().Value)
	})
	if err != nil {
		results = append(results, framework.ResultItem{Message: fmt.Sprintf("Resource %q encountered error %q", object.GetName(), err.Error()), Severity: framework.Error})
	}
	results = append(results, fw.Results()...)

	annoString, err := fw.Annotation()
	if err != nil {
		return results, err
	}
	if annoString != "" {
		currentAnno := object.GetAnnotations()
		currentAnno[annotationKey] = annoString
		err = object.SetAnnotations(currentAnno)
		if err != nil {
			return results, err
		}
	}

	return results, err
}

func (rp *AnnotateApplyTimeMutationsProcessor) Process(resourceList *framework.ResourceList) error {
	resourceList.Result = &framework.Result{
		Name: "annotate-apply-time-mutations",
	}
	for _, node := range resourceList.Items {
		fileName, _, err := kioutil.GetFileAnnotations(node)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Processing resource in file", fileName)
		results, err := processResource(node, fileName)
		if err != nil {
			return err
		}
		resourceList.Result.Items = append(resourceList.Result.Items, results...)
	}

	return nil
}

func main() {
	rp := AnnotateApplyTimeMutationsProcessor{}
	cmd := command.Build(&rp, command.StandaloneEnabled, false)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/annotate-mutations/annotatemutations"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	mutationCommentIdentifier = "# apply-time-mutation: "
	annotationKey             = "config.kubernetes.io/apply-time-mutation"
)

type mutatorAnnotation []annotatemutations.Mutation

type commentProcessor struct {
	results []framework.ResultItem
}

type fieldWalker struct {
	parentProcessor *commentProcessor
	fileName        string
	// mutationCount increments as field mutations are found, to ensure unique replacement tokens in multi-comment resources.
	mutationCount int
	annotation    mutatorAnnotation
}

// extractMutationPattern extracts the setter pattern from the line comment.
// If the the line comment doesn't contain MutationCommentIdentifier
// prefix, then it returns an empty string.
func extractMutationPattern(lineComment string) string {
	if !strings.HasPrefix(lineComment, mutationCommentIdentifier) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(lineComment, mutationCommentIdentifier))
}

func (fw *fieldWalker) visitField(node *yaml.RNode, n string) error {
	// Visit fields with comments.
	if comment := node.YNode().LineComment; comment != "" {
		fmt.Fprintln(os.Stderr, "Parsing field comment", comment)
		// Check if comment is a mutation annotation.
		mutationPattern := extractMutationPattern(comment)
		if mutationPattern == "" {
			// The comment is not a mutation annotation.
			return nil
		}
		if !annotatemutations.HasRef(mutationPattern) {
			// Mutation comment is tagged but no valid reference found.
			return fmt.Errorf("apply mutation comment found with no valid reference to source path")
		}

		resourceRef, refPath := annotatemutations.CommentToReference(mutationPattern)

		// replace the setter names in comment pattern with provided values
		replacementValue, replacementToken := annotatemutations.CommentToTokenField(mutationPattern, fw.mutationCount)
		fw.mutationCount++
		if replacementValue != "" {
			node.YNode().Value = replacementValue
		}

		fw.annotation = append(fw.annotation, annotatemutations.Mutation{SourceRef: resourceRef, SourcePath: refPath, TargetPath: "$." + n, Token: replacementToken})
		fw.parentProcessor.results = append(fw.parentProcessor.results, framework.ResultItem{Message: fmt.Sprintf("Parsed mutation in resource %q field %q", fw.fileName, n), Severity: framework.Info})
	}
	return nil
}

// visitKeys recurses over a complex yaml hierarchy
func (fw *fieldWalker) visitKeys(object *yaml.RNode, p string) error {
	switch object.YNode().Kind {
	case yaml.MappingNode:
		// iterate over map values
		return object.VisitFields(func(node *yaml.MapNode) error {
			return fw.visitKeys(node.Value, p+"."+node.Key.YNode().Value)
		})
	case yaml.SequenceNode:
		els, err := object.Elements()
		if err != nil {
			return err
		}
		// iterate over list elements
		for i, field := range els {
			err := fw.visitKeys(field, p+fmt.Sprintf("[%d]", i))
			if err != nil {
				return err
			}
		}
		return nil
	case yaml.ScalarNode:
		// visit the scalar field
		return fw.visitField(object, p)
	}
	return nil
}

func (rp *commentProcessor) visitResource(object *yaml.RNode, resourcePath string) error {
	if object.IsNil() {
		return nil
	}

	// one fw instance per resource
	fw := fieldWalker{parentProcessor: rp, fileName: resourcePath}
	err := object.VisitFields(func(node *yaml.MapNode) error {
		return fw.visitKeys(node.Value, node.Key.YNode().Value)
	})
	if err != nil {
		rp.results = append(rp.results, framework.ResultItem{Message: fmt.Sprintf("Resource %q encountered error %q", object.GetName(), err.Error()), Severity: framework.Error})
	}

	if len(fw.annotation) > 0 {
		serialized, err := yaml.Marshal(fw.annotation)
		if err != nil {
			return err
		}
		currentAnno := object.GetAnnotations()
		currentAnno[annotationKey] = string(serialized)
		err = object.SetAnnotations(currentAnno)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rp *commentProcessor) Process(resourceList *framework.ResourceList) error {
	for _, node := range resourceList.Items {
		fileName, _, err := kioutil.GetFileAnnotations(node)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Processing resource in file", fileName)
		err = rp.visitResource(node, fileName)
		if err != nil {
			return err
		}
	}
	resourceList.Result = &framework.Result{
		Name: "annotate-mutations",
	}
	resourceList.Result.Items = rp.results
	return nil
}

func main() {
	rp := commentProcessor{}
	cmd := command.Build(&rp, command.StandaloneEnabled, false)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

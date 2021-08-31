package annotateapplytimemutations

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	tokenPattern              = "$ref%d"
	sourceRegex               = `\${(?P<group>[^/]+)/((?P<version>[^/]+)/)?namespaces/(?P<namespace>[^/]+)/(?P<kind>[^/]+)/(?P<name>[^:]+):(?P<path>[^}]+)}`
	onlySourceRegex           = "^" + sourceRegex + "$"
	mutationCommentIdentifier = "# apply-time-mutation: "
	annotationKey             = "config.kubernetes.io/apply-time-mutation"
)

var (
	resourceReferencePattern = regexp.MustCompile(sourceRegex)
	onlyReferencePattern     = regexp.MustCompile(onlySourceRegex)
)

type RefStruct struct {
	Group      string `yaml:"group,omitempty"`
	ApiVersion string `yaml:"apiVersion,omitempty"`
	Kind       string `yaml:"kind"`
	Name       string `yaml:"name"`
	Namespace  string `yaml:"namespace"`
}

type Mutation struct {
	SourceRef  RefStruct `yaml:"sourceRef"`
	SourcePath string    `yaml:"sourcePath"`
	TargetPath string    `yaml:"targetPath"`
	Token      string    `yaml:"token,omitempty"`
}

type mutatorAnnotation []Mutation

// ResourceAnnotator scans all fields in a resource for mutation annotation comments and adds the matching annotation.
type ResourceAnnotator struct {
	// fileName is the name of the file or path where the object to annotate is located.
	fileName string
	results  []framework.ResultItem
	// mutationCount increments as field mutations are found, to ensure unique replacement tokens in multi-comment resources.
	mutationCount int
	annotation    mutatorAnnotation
}

// annotationAsString generates the krm annotation string value, or errors.
func (fw *ResourceAnnotator) annotationAsString() (string, error) {
	if len(fw.annotation) > 0 {
		serialized, err := yaml.Marshal(fw.annotation)
		if err != nil {
			return "", err
		}
		return string(serialized), nil
	}
	return "", nil
}

// hasRef returns whether or not the comment has a source reference embeded.
func hasRef(comment string) bool {
	return resourceReferencePattern.MatchString(comment)
}

// commentToReference parses a comment source reference to return the structured annotation fields.
func commentToReference(comment string) (RefStruct, string) {
	outs := resourceReferencePattern.FindStringSubmatch(comment)
	group := outs[1]
	version := outs[3]
	retVal := RefStruct{
		Namespace: outs[4],
		Kind:      outs[5],
		Name:      outs[6],
	}
	if version != "" {
		retVal.ApiVersion = fmt.Sprintf("%s/%s", group, version)
	} else {
		retVal.Group = group
	}
	return retVal, outs[7]
}

// commentToTokenField replaces source reference strings with a replacement token.
// Returns replaced tokenized field value, and the replacement token to reference in the annotation.
func commentToTokenField(comment string, index int) (string, string) {
	// If the mutation comment is *only* a source reference, do not tokenize.
	if onlyReferencePattern.MatchString(comment) {
		return "", ""
	}
	token := fmt.Sprintf(tokenPattern, index)
	return resourceReferencePattern.ReplaceAllLiteralString(comment, token), token
}

// extractMutationPattern extracts the mutation pattern from the line comment.
// If the the line comment doesn't contain MutationCommentIdentifier
// prefix, then it returns an empty string.
func extractMutationPattern(lineComment string) string {
	if !strings.HasPrefix(lineComment, mutationCommentIdentifier) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(lineComment, mutationCommentIdentifier))
}

// visitScalarNode searches for mutation markup comments and parses them to the equivalent annotations
func (fw *ResourceAnnotator) visitScalarNode(node *yaml.RNode, n string) error {
	// Visit fields with comments.
	if comment := node.YNode().LineComment; comment != "" {
		fmt.Fprintln(os.Stderr, "Parsing field comment", comment)
		// Check if comment is a mutation annotation.
		mutationPattern := extractMutationPattern(comment)
		if mutationPattern == "" {
			// The comment is not a mutation annotation.
			return nil
		}
		if !hasRef(mutationPattern) {
			// Mutation comment is tagged but no valid reference found.
			return fmt.Errorf("apply mutation comment found with no valid reference to source path")
		}

		resourceRef, refPath := commentToReference(mutationPattern)

		// replace the setter names in comment pattern with provided values
		replacementValue, replacementToken := commentToTokenField(mutationPattern, fw.mutationCount)
		fw.mutationCount++
		if replacementValue != "" {
			node.YNode().Value = replacementValue
		}

		fw.annotation = append(fw.annotation, Mutation{SourceRef: resourceRef, SourcePath: refPath, TargetPath: "$." + n, Token: replacementToken})
		fw.results = append(fw.results, framework.ResultItem{Message: fmt.Sprintf("Parsed mutation in resource %q field %q", fw.fileName, n), Severity: framework.Info})
	}
	return nil
}

// visitFields recurses over a yaml map of arbitrary complexity.
// This is the entry point for processing any krm object.
func (fw *ResourceAnnotator) visitFields(object *yaml.RNode, p string) error {
	switch object.YNode().Kind {
	case yaml.MappingNode:
		// iterate over map values
		return object.VisitFields(func(node *yaml.MapNode) error {
			return fw.visitFields(node.Value, p+"."+node.Key.YNode().Value)
		})
	case yaml.SequenceNode:
		els, err := object.Elements()
		if err != nil {
			return err
		}
		// iterate over list elements
		for i, field := range els {
			err := fw.visitFields(field, p+fmt.Sprintf("[%d]", i))
			if err != nil {
				return err
			}
		}
		return nil
	case yaml.ScalarNode:
		// visit the scalar node
		return fw.visitScalarNode(object, p)
	}
	return nil
}

// AnnotateResource parses comments on fields in one resource and adds the corresponding annotations.
func (fw *ResourceAnnotator) AnnotateResource(object *yaml.RNode, filePath string) ([]framework.ResultItem, error) {
	var results []framework.ResultItem
	fw.fileName = filePath
	err := object.VisitFields(func(node *yaml.MapNode) error {
		return fw.visitFields(node.Value, node.Key.YNode().Value)
	})
	if err != nil {
		results = append(results, framework.ResultItem{Message: fmt.Sprintf("Resource %q encountered an error: %q", object.GetName(), err.Error()), Severity: framework.Error})
		return results, err
	}
	results = fw.results

	annoString, err := fw.annotationAsString()
	if err != nil {
		results = append(results, framework.ResultItem{Message: fmt.Sprintf("Resource %q encountered an error serializing the mutation annotation: %q", object.GetName(), err.Error()), Severity: framework.Error})
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

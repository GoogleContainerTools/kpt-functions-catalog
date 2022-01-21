package pkg

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"sigs.k8s.io/cli-utils/pkg/object"
	"sigs.k8s.io/cli-utils/pkg/object/mutation"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	tokenPattern              = "${ref%d}"
	sourceRegex               = `\${(?P<group>[^/]+)/((?P<version>[^/]+)/)?namespaces/(?P<namespace>[^/]+)/(?P<kind>[^/]+)/(?P<name>[^:]+):(?P<path>[^}]+)}`
	onlySourceRegex           = "^" + sourceRegex + "$"
	mutationCommentIdentifier = "# apply-time-mutation: "
	ResourceGroup             = "fn.kpt.dev"
	ResourceVersion           = "v1alpha1"
)

var (
	resourceReferencePattern = regexp.MustCompile(sourceRegex)
	onlyReferencePattern     = regexp.MustCompile(onlySourceRegex)
)

type CommentScanner struct {
	// ObjMeta is a reference to the object using metadata
	ObjMeta yaml.ResourceIdentifier
	// ObjFile is a reference to the object using file and object index
	ObjFile framework.File
}

type ScanResult struct {
	Path         string
	Value        interface{}
	Comment      string
	Substitution mutation.FieldSubstitution
}

// Scan searches for mutation markup comments and parses them as substitutions.
func (cs *CommentScanner) Scan(obj *yaml.RNode, results map[string]ScanResult, fieldPath ...interface{}) error {
	switch obj.YNode().Kind {
	case yaml.MappingNode:
		// iterate over map (key->value)
		err := obj.VisitFields(func(node *yaml.MapNode) error {
			key, err := nodeValue(node.Key)
			if err != nil {
				return fmt.Errorf("invalid map key %q (path: %q): %w", node.Key.YNode().Value, fieldPath, err)
			}
			return cs.Scan(node.Value, results, append(fieldPath, key)...)
		})
		return err
	case yaml.SequenceNode:
		// iterate over sequence (index->value)
		// Can't use VisitElements, because it doesn't provide the index...
		els, err := obj.Elements()
		if err != nil {
			return fmt.Errorf("invalid sequence (path: %q): %w", fieldPath, err)
		}
		for i, field := range els {
			err := cs.Scan(field, results, append(fieldPath, i)...)
			if err != nil {
				return err
			}
		}
		return nil
	case yaml.ScalarNode:
		// scan the scalar node
		return cs.scanScalar(obj, results, fieldPath...)
	}
	return nil
}

func (cs *CommentScanner) scanScalar(node *yaml.RNode, results map[string]ScanResult, fieldPath ...interface{}) error {
	comment := node.YNode().LineComment
	if comment == "" {
		// empty or missing comment
		return nil
	}

	// format field path with JSONPath
	// TODO: will this break kyaml fn framework? kyaml paths seem to be just maps fields and just period delimiters...
	fieldPathStr := strings.TrimPrefix(object.FieldPath(fieldPath), ".")

	fmt.Fprintf(os.Stderr, "Parsing comment (field: %q): %s\n", fieldPathStr, comment)
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

	currentValue, err := nodeValue(node)
	if err != nil {
		return fmt.Errorf("invalid node value (path: %q): %w", fieldPathStr, err)
	}

	// replace the setter names in comment pattern with provided values
	replacementValue, replacementToken := commentToTokenField(mutationPattern, len(results))
	if replacementValue != "" {
		node.YNode().Value = replacementValue
	}

	results[fieldPathStr] = ScanResult{
		Path:    fieldPathStr,
		Value:   currentValue,
		Comment: comment,
		Substitution: mutation.FieldSubstitution{
			SourceRef:  resourceRef,
			SourcePath: refPath,
			TargetPath: "$." + fieldPathStr,
			Token:      replacementToken,
		},
	}
	return nil
}

// hasRef returns whether or not the comment has a source reference embeded.
func hasRef(comment string) bool {
	return resourceReferencePattern.MatchString(comment)
}

// commentToReference parses a comment source reference to return the structured annotation fields.
func commentToReference(comment string) (mutation.ResourceReference, string) {
	outs := resourceReferencePattern.FindStringSubmatch(comment)
	group := outs[1]
	version := outs[3]
	retVal := mutation.ResourceReference{
		Namespace: outs[4],
		Kind:      outs[5],
		Name:      outs[6],
	}
	if version != "" {
		retVal.APIVersion = fmt.Sprintf("%s/%s", group, version)
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

// nodeValue decodes a YAML node value of any type.
// Returns an error if the value is unparsable/invalid.
func nodeValue(node *yaml.RNode) (interface{}, error) {
	var value interface{}
	err := node.YNode().Decode(&value)
	if err != nil {
		return value, fmt.Errorf("failed to decode field value: %w", err)
	}
	return value, nil
}

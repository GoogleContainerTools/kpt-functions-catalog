package annotatemutations

import (
	"fmt"
	"regexp"
)

const (
	tokenPattern    = "$ref%d"
	sourceRegex     = `\${(?P<group>[^/]+)/((?P<version>[^/]+)/)?namespaces/(?P<namespace>[^/]+)/(?P<kind>[^/]+)/(?P<name>[^:]+):(?P<path>[^}]+)}`
	onlySourceRegex = "^" + sourceRegex + "$"
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

// HasRef returns whether or not the comment has a source reference embeded.
func HasRef(comment string) bool {
	return resourceReferencePattern.MatchString(comment)
}

// CommentToReference parses a comment source reference to return the structured annotation fields.
func CommentToReference(comment string) (RefStruct, string) {
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

// CommentToTokenField replaces source reference strings with a replacement token.
// Returns replaced tokenized field value, and the replacement token to reference in the annotation.
func CommentToTokenField(comment string, index int) (string, string) {
	// If the mutation comment is *only* a source reference, do not tokenize.
	if onlyReferencePattern.MatchString(comment) {
		return "", ""
	}
	token := fmt.Sprintf(tokenPattern, index)
	return resourceReferencePattern.ReplaceAllLiteralString(comment, token), token
}

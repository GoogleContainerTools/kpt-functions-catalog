package image_util

import (
	"regexp"
	"strings"
)

// Image contains an image name, a new name, a new tag or digest,
// which will replace the original name and tag.
type Image struct {
	// Name is a tag-less image name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// NewName is the value used to replace the original name.
	NewName string `json:"newName,omitempty" yaml:"newName,omitempty"`

	// NewTag is the value used to replace the original tag.
	NewTag string `json:"newTag,omitempty" yaml:"newTag,omitempty"`

	// Digest is the value used to replace the original image tag.
	// If digest is present NewTag value is ignored.
	Digest string `json:"digest,omitempty" yaml:"digest,omitempty"`
}

type FsSlice []FieldSpec

type FieldSpec struct {
	Gvk                `json:",inline,omitempty" yaml:",inline,omitempty"`
	Path               string `json:"path,omitempty" yaml:"path,omitempty"`
	CreateIfNotPresent bool   `json:"create,omitempty" yaml:"create,omitempty"`
}

type Gvk struct {
	Group   string `json:"group,omitempty" yaml:"group,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	Kind    string `json:"kind,omitempty" yaml:"kind,omitempty"`
}

var pattern *regexp.Regexp

// IsImageMatched returns true if the value of t is identical to the
// image name in the full image name and tag as given by s.
func IsImageMatched(s, t string) bool {
	// Tag values are limited to [a-zA-Z0-9_.{}-].
	// Some tools like Bazel rules_k8s allow tag patterns with {} characters.
	// More info: https://github.com/bazelbuild/rules_k8s/pull/423
	pattern, _ = regexp.Compile("^" + t + "(@sha256)?(:[a-zA-Z0-9_.{}-]*)?$")
	return pattern.MatchString(s)
}

// Split separates and returns the name and tag parts
// from the image string using either colon `:` or at `@` separators.
// Note that the returned tag keeps its separator.
func Split(imageName string) (name string, tag string) {
	// check if image name contains a domain
	// if domain is present, ignore domain and check for `:`
	ic := -1
	if slashIndex := strings.Index(imageName, "/"); slashIndex < 0 {
		ic = strings.LastIndex(imageName, ":")
	} else {
		lastIc := strings.LastIndex(imageName[slashIndex:], ":")
		// set ic only if `:` is present
		if lastIc > 0 {
			ic = slashIndex + lastIc
		}
	}
	ia := strings.LastIndex(imageName, "@")
	if ic < 0 && ia < 0 {
		return imageName, ""
	}

	i := ic
	if ia > 0 {
		i = ia
	}

	name = imageName[:i]
	tag = imageName[i:]
	return
}

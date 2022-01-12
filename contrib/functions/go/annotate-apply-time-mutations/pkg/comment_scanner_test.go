package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/cli-utils/pkg/object/mutation"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestApplySettersReferenceParse(t *testing.T) {
	testCases := []struct {
		s          string
		wantStruct mutation.ResourceReference
		wantPath   string
	}{
		{
			s: "${foo.bar.com/namespaces/example-namespace/aKind/example-name:$.my.field}",
			wantStruct: mutation.ResourceReference{
				Group:     "foo.bar.com",
				Kind:      "aKind",
				Name:      "example-name",
				Namespace: "example-namespace",
			},
			wantPath: "$.my.field",
		},
		{
			s: "${foo.bar.com/v1alpha1/namespaces/example-namespace/aKind/example-name:$.my.field}",
			wantStruct: mutation.ResourceReference{
				APIVersion: "foo.bar.com/v1alpha1",
				Kind:       "aKind",
				Name:       "example-name",
				Namespace:  "example-namespace",
			},
			wantPath: "$.my.field",
		},
	}

	for _, test := range testCases {
		t.Run(test.s, func(t *testing.T) {
			doesHaveRef := hasRef(test.s)
			if !doesHaveRef {
				t.Fatalf("String %q doesn't have a valid ref", test.s)
			}

			gotRef, gotPath := commentToReference(test.s)

			if gotRef != test.wantStruct {
				t.Errorf("CommentToReference returned struct %v wanted %v", gotRef, test.wantStruct)
			}

			if gotPath != test.wantPath {
				t.Errorf("CommentToReference returned path %q wanted %q", gotPath, test.wantPath)
			}
		})
	}
}

func TestCommentToTokenField(t *testing.T) {
	testCases := []struct {
		s          string
		givenIndex int
		wantValue  string
		wantToken  string
	}{
		{
			s:          "prefix-${foo.bar.com/namespaces/example-namespace/aKind/example-name:$.my.field}-suffix",
			givenIndex: 5,
			wantValue:  "prefix-${ref5}-suffix",
			wantToken:  "${ref5}",
		},
		{
			s:          "${foo.bar.com/v1alpha1/namespaces/example-namespace/aKind/example-name:$.my.field}",
			givenIndex: 2,
			wantValue:  "",
			wantToken:  "",
		},
	}

	for _, test := range testCases {
		t.Run(test.s, func(t *testing.T) {
			gotReplace, gotToken := commentToTokenField(test.s, test.givenIndex)
			if gotReplace != test.wantValue {
				t.Errorf("CommentToTokenField returned replacement %q want %q", gotReplace, test.wantValue)
			}
			if gotToken != test.wantToken {
				t.Errorf("CommentToTokenField returned token %q want %q", gotToken, test.wantToken)
			}
		})
	}
}

func TestCommentScan(t *testing.T) {
	testCases := []struct {
		config        string
		expectResults map[string]ScanResult
	}{
		{
			config: `apiVersion: bar.foo/v1beta1
kind: MyTestKind
metadata:
    name: my-test-resource
    namespace: test-namespace
    annotations:
        unmodified-key: foobarbaz
spec:
    a: 0 # apply-time-mutation: ${foo.bar/v0/namespaces/example-namespace/OtherKind/example-name2:$.status.count}
`,
			expectResults: map[string]ScanResult{
				"spec.a": {
					Path:    "spec.a",
					Value:   0,
					Comment: "# apply-time-mutation: ${foo.bar/v0/namespaces/example-namespace/OtherKind/example-name2:$.status.count}",
					Substitution: mutation.FieldSubstitution{
						SourceRef: mutation.ResourceReference{
							APIVersion: "foo.bar/v0",
							Kind:       "OtherKind",
							Name:       "example-name2",
							Namespace:  "example-namespace",
						},
						SourcePath: "$.status.count",
						TargetPath: "$.spec.a",
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run("", func(t *testing.T) {
			node, err := kyaml.Parse(test.config)
			assert.NoError(t, err)
			meta, err := node.GetMeta()
			assert.NoError(t, err)
			scanner := CommentScanner{
				ObjMeta: meta.GetIdentifier(),
				ObjFile: framework.File{
					Path:  "test.yaml",
					Index: 0,
				},
			}
			results := make(map[string]ScanResult)
			err = scanner.Scan(node, results)
			assert.NoError(t, err)
			assert.Equal(t, test.expectResults, results)
		})
	}
}

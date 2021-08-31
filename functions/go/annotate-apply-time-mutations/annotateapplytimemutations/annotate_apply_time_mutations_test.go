package annotateapplytimemutations

import (
	"fmt"
	"testing"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestApplySettersReferenceParse(t *testing.T) {
	testCases := []struct {
		s          string
		wantStruct RefStruct
		wantPath   string
	}{
		{
			s: "${foo.bar.com/namespaces/example-namespace/aKind/example-name:$.my.field}",
			wantStruct: RefStruct{
				Group:     "foo.bar.com",
				Kind:      "aKind",
				Name:      "example-name",
				Namespace: "example-namespace",
			},
			wantPath: "$.my.field",
		},
		{
			s: "${foo.bar.com/v1alpha1/namespaces/example-namespace/aKind/example-name:$.my.field}",
			wantStruct: RefStruct{
				ApiVersion: "foo.bar.com/v1alpha1",
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
			wantValue:  "prefix-$ref5-suffix",
			wantToken:  "$ref5",
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

func TestAnnotateResourceOutput(t *testing.T) {
	testCases := []struct {
		config            string
		expectAnnotations map[string]string
		expectResults     []framework.ResultItem
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
			expectAnnotations: map[string]string{
				"config.kubernetes.io/apply-time-mutation": `- sourceRef:
    apiVersion: foo.bar/v0
    kind: OtherKind
    name: example-name2
    namespace: example-namespace
  sourcePath: $.status.count
  targetPath: $.spec.a
`,
				"unmodified-key": "foobarbaz",
			},
			expectResults: []framework.ResultItem{
				{Message: fmt.Sprintf("Parsed mutation in file %q field %q", "test.yaml", "spec.a"), Severity: framework.Info},
			},
		},
	}

	for _, test := range testCases {
		t.Run("", func(t *testing.T) {
			node, err := kyaml.Parse(test.config)
			if err != nil {
				t.Fatal(err)
			}
			ra := ResourceAnnotator{}
			gotResults, err := ra.AnnotateResource(node, "test.yaml")
			gotAnnotations := node.GetAnnotations()

			if len(gotResults) != len(test.expectResults) {
				t.Fatalf("Got %d results expected %d", len(gotResults), len(test.expectResults))
			}
			if len(gotAnnotations) != len(test.expectAnnotations) {
				t.Fatalf("Got %d annotations expected %d", len(gotAnnotations), len(test.expectAnnotations))
			}
			for i, gotResult := range gotResults {
				if gotResult != test.expectResults[i] {
					t.Errorf("Got %dth result %v, expected %v", i, gotResult, test.expectResults[i])
				}
			}
			for gotKey, gotVal := range gotAnnotations {
				if gotVal != test.expectAnnotations[gotKey] {
					t.Errorf("Got %q: %q, expected value: %q", gotKey, gotVal, test.expectAnnotations[gotKey])
				}
			}
		})
	}
}

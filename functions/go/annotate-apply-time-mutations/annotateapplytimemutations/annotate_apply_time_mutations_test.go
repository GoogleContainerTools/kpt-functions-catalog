package annotateapplytimemutations

import "testing"

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
			doesHaveRef := HasRef(test.s)
			if !doesHaveRef {
				t.Fatalf("String %q doesn't have a valid ref", test.s)
			}

			gotRef, gotPath := CommentToReference(test.s)

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
			gotReplace, gotToken := CommentToTokenField(test.s, test.givenIndex)
			if gotReplace != test.wantValue {
				t.Errorf("CommentToTokenField returned replacement %q want %q", gotReplace, test.wantValue)
			}
			if gotToken != test.wantToken {
				t.Errorf("CommentToTokenField returned token %q want %q", gotToken, test.wantToken)
			}
		})
	}
}

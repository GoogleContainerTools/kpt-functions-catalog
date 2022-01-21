package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cli-utils/pkg/object/mutation"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestObjectScan(t *testing.T) {
	testCases := []struct {
		config       string
		expectResult *ApplyTimeMutation
	}{
		{
			config: `apiVersion: fn.kpt.dev/v1alpha1
kind: ApplyTimeMutation
metadata:
  name: example
spec:
  targetRef:
    kind: ConfigMap
    name: target-object
    namespace: test-namespace
  substitutions:
  - sourceRef:
      kind: ConfigMap
      name: source-object
      namespace: test-namespace
    sourcePath: $.spec.data
    targetPath: $.spec.data
`,
			expectResult: &ApplyTimeMutation{
				TypeMeta: v1.TypeMeta{
					APIVersion: "fn.kpt.dev/v1alpha1",
					Kind:       "ApplyTimeMutation",
				},
				ObjectMeta: v1.ObjectMeta{
					Name: "example",
				},
				Spec: ApplyTimeMutationSpec{
					TargetRef: mutation.ResourceReference{
						Kind:      "ConfigMap",
						Name:      "target-object",
						Namespace: "test-namespace",
					},
					Substitutions: []mutation.FieldSubstitution{
						{
							SourceRef: mutation.ResourceReference{
								Kind:      "ConfigMap",
								Name:      "source-object",
								Namespace: "test-namespace",
							},
							SourcePath: "$.spec.data",
							TargetPath: "$.spec.data",
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run("", func(t *testing.T) {
			node, err := kyaml.Parse(test.config)
			assert.NoError(t, err)
			scanner := ObjectScanner{}
			atm, err := scanner.Scan(node)
			assert.NoError(t, err)
			assert.Equal(t, test.expectResult, atm)
		})
	}
}

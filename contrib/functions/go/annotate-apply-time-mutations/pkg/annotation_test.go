package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/cli-utils/pkg/object/mutation"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestWriteAnnotation(t *testing.T) {
	testCases := []struct {
		config         string
		subs           mutation.ApplyTimeMutation
		expectedResult string
	}{
		{
			config: `apiVersion: bar.foo/v1beta1
kind: MyTestKind
metadata:
  name: my-test-resource
  namespace: test-namespace
spec: {}
`,
			subs: mutation.ApplyTimeMutation{
				mutation.FieldSubstitution{
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
			expectedResult: `apiVersion: bar.foo/v1beta1
kind: MyTestKind
metadata:
  name: my-test-resource
  namespace: test-namespace
  annotations:
    config.kubernetes.io/apply-time-mutation: |
      - sourcePath: $.status.count
        sourceRef:
          apiVersion: foo.bar/v0
          kind: OtherKind
          name: example-name2
          namespace: example-namespace
        targetPath: $.spec.a
spec: {}
`,
		},
	}

	for _, test := range testCases {
		t.Run("", func(t *testing.T) {
			node, err := yaml.Parse(test.config)
			assert.NoError(t, err)

			err = WriteAnnotation(node, test.subs)
			assert.NoError(t, err)

			result, err := node.String()
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResult, result)
		})
	}
}

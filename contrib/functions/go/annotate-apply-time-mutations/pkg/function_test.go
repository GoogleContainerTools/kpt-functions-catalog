package pkg

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/cli-utils/pkg/testutil"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestFunctionProcess(t *testing.T) {
	testCases := map[string]struct {
		configs         [][]string
		expectedConfigs [][]string
		expectedResults framework.Results
	}{
		"apply-time-mutation comment": {
			configs: [][]string{
				{
					`apiVersion: bar.foo/v1beta1
kind: MyTestKind
metadata:
  name: my-test-resource
  namespace: test-namespace
spec:
  a: 0 # apply-time-mutation: ${foo.bar/v0/namespaces/example-namespace/OtherKind/example-name2:$.status.count}
`,
				},
			},
			expectedConfigs: [][]string{
				{
					`apiVersion: bar.foo/v1beta1
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
spec:
  a: 0 # apply-time-mutation: ${foo.bar/v0/namespaces/example-namespace/OtherKind/example-name2:$.status.count}
`,
				},
			},
			expectedResults: framework.Results{
				&framework.Result{
					Message:  "Found valid apply-time-mutation comment (item: 0): # apply-time-mutation: ${foo.bar/v0/namespaces/example-namespace/OtherKind/example-name2:$.status.count}",
					Severity: "info",
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "bar.foo/v1beta1",
							Kind:       "MyTestKind",
						},
						NameMeta: yaml.NameMeta{
							Name:      "my-test-resource",
							Namespace: "test-namespace",
						},
					},
					Field: &framework.Field{Path: "$.spec.a", CurrentValue: 0, ProposedValue: nil},
					File:  &framework.File{Path: "file-0.yaml", Index: 0},
				},
				&framework.Result{
					Message:  "Wrote apply-time-mutation annotation (item: 0)",
					Severity: "info",
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "bar.foo/v1beta1",
							Kind:       "MyTestKind",
						},
						NameMeta: yaml.NameMeta{
							Name:      "my-test-resource",
							Namespace: "test-namespace",
						},
					},
					File: &framework.File{Path: "file-0.yaml", Index: 0},
				},
			},
		},
		"ApplyTimeMutation object": {
			configs: [][]string{
				{
					`apiVersion: fn.kpt.dev/v1alpha1
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
				},
				{
					`apiVersion: v1
kind: ConfigMap
metadata:
  name: source-object
  namespace: test-namespace
spec:
  data:
    foo: hello
    bar: world
`,
					`apiVersion: v1
kind: ConfigMap
metadata:
  name: target-object
  namespace: test-namespace
spec:
  data: {}
`,
				},
			},
			expectedConfigs: [][]string{
				{
					`apiVersion: fn.kpt.dev/v1alpha1
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
				},
				{
					`apiVersion: v1
kind: ConfigMap
metadata:
  name: source-object
  namespace: test-namespace
spec:
  data:
    foo: hello
    bar: world
`,
					`apiVersion: v1
kind: ConfigMap
metadata:
  name: target-object
  namespace: test-namespace
  annotations:
    config.kubernetes.io/apply-time-mutation: |
      - sourcePath: $.spec.data
        sourceRef:
          kind: ConfigMap
          name: source-object
          namespace: test-namespace
        targetPath: $.spec.data
spec:
  data: {}
`,
				},
			},
			expectedResults: framework.Results{
				&framework.Result{
					Message:  "Found valid ApplyTimeMutation object (item: 0)",
					Severity: "info",
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "fn.kpt.dev/v1alpha1",
							Kind:       "ApplyTimeMutation",
						},
						NameMeta: yaml.NameMeta{
							Name: "example",
						},
					},
					File: &framework.File{Path: "file-0.yaml", Index: 0},
				},
				&framework.Result{
					Message:  "Wrote apply-time-mutation annotation (item: 2)",
					Severity: "info",
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "ConfigMap",
						},
						NameMeta: yaml.NameMeta{
							Name:      "target-object",
							Namespace: "test-namespace",
						},
					},
					File: &framework.File{Path: "file-1.yaml", Index: 1},
				},
			},
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			items, err := stringsToItems(test.configs)
			assert.NoError(t, err)

			resourceList := &framework.ResourceList{
				Items: items,
			}

			fn := Function{}
			err = fn.Process(resourceList)
			assert.NoError(t, err)

			assert.Equal(t, test.expectedResults, resourceList.Results)

			configs, err := itemsToStrings(resourceList.Items)
			assert.NoError(t, err)
			testutil.AssertEqual(t, test.expectedConfigs, configs)
		})
	}
}

// stringsToItems simulates reading a package of yaml files, each with zero or
// more objects.
func stringsToItems(configs [][]string) ([]*yaml.RNode, error) {
	var items []*yaml.RNode
	for i, file := range configs {
		for j, config := range file {
			node, err := yaml.Parse(config)
			if err != nil {
				return items, fmt.Errorf("failed to parse yaml (configs[%d][%d]): %v", i, j, err)
			}
			a := node.GetAnnotations()
			a[kioutil.PathAnnotation] = fmt.Sprintf("file-%d.yaml", i)
			a[kioutil.IndexAnnotation] = strconv.Itoa(j)
			err = node.SetAnnotations(a)
			if err != nil {
				return items, fmt.Errorf("failed to update annotations: %v", err)
			}
			items = append(items, node)
		}
	}
	return items, nil
}

// itemsToStrings simulates writing a package of yaml files, each with zero or
// more objects.
func itemsToStrings(items []*yaml.RNode) ([][]string, error) {
	filePathPattern := regexp.MustCompile(`file-(\d+).yaml`)
	var configs [][]string
	for i, item := range items {
		a := item.GetAnnotations()
		filePath := a[kioutil.PathAnnotation]
		fileIndexStr := a[kioutil.IndexAnnotation]
		delete(a, kioutil.PathAnnotation)
		delete(a, kioutil.IndexAnnotation)
		err := item.SetAnnotations(a)
		if err != nil {
			return configs, fmt.Errorf("failed to update annotations: %v", err)
		}

		match := filePathPattern.FindStringSubmatch(filePath)
		if len(match) != 2 {
			return configs, fmt.Errorf("failed to parse file path (item: %d): %q", i, filePath)
		}
		filePathIndex, err := strconv.Atoi(match[1])
		if err != nil {
			return configs, fmt.Errorf("failed to parse file path (item: %d): %q: %w", i, filePath, err)
		}

		fileIndex, err := strconv.Atoi(fileIndexStr)
		if err != nil {
			return configs, fmt.Errorf("failed to parse file index (item: %d): %q: %w", i, fileIndexStr, err)
		}

		config, err := item.String()
		if err != nil {
			return configs, fmt.Errorf("failed to format yaml (configs[%d][%d]): %v", filePathIndex, fileIndex, err)
		}

		for filePathIndex >= len(configs) {
			// grow file list
			configs = append(configs, []string{})
		}
		for fileIndex >= len(configs[filePathIndex]) {
			// grow object list
			configs[filePathIndex] = append(configs[filePathIndex], "")
		}
		configs[filePathIndex][fileIndex] = config
	}
	return configs, nil
}

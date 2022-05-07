package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"github.com/stretchr/testify/assert"
)

// helper function to convert ResourceList items to yaml
func itemsToYaml(items []*fn.KubeObject) string {
	var itemYamls []string
	for _, item := range items {
		itemYamls = append(itemYamls, item.String())
	}
	result := strings.Join(itemYamls, "---\n")
	return result
}

func runImageTransformerResults(input, config string) (*fn.ResourceList, error) {
	rl, err := fn.ParseResourceList([]byte(input))
	if err != nil {
		return nil, err
	}
	functionConfig, err := fn.ParseKubeObject([]byte(config))
	if err != nil {
		return nil, err
	}
	rl.FunctionConfig = functionConfig
	in, _ := rl.ToYAML()
	out, err := fn.Run(fn.ResourceListProcessorFunc(setImageTags), in)
	if err != nil {
		return nil, err
	}
	rl, _ = fn.ParseResourceList(out)

	ko, _ := fn.ParseKubeObject(out)
	results := ko.GetSlice("results")
	for _, result := range results{
		if result.GetString("severity") == "error" {
			return rl, fn.GeneralResult(result.GetString("message"), fn.Error)
		}
		rl.Results = append(rl.Results, fn.GeneralResult(result.GetString("message"), fn.Info))
	}
	return rl, nil
}

func runImageTransformerE(input, config string) (string, error) {
	rl, err := runImageTransformerResults(input, config)
	if err != nil {
		return "", err
	}
	rl.Sort()
	result := itemsToYaml(rl.Items)
	return result, nil
}

func runImageTransformer(t *testing.T, input, config string) string {
	s, err := runImageTransformerE(input, config)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestImageTransformer(t *testing.T) {
	testCases := []struct {
		TestName       string
		FunctionConfig string
		Input          string
		ExpectedOutput string
	}{
		{
			TestName: "set-image should accept a ConfigMap functionConfig",
			FunctionConfig: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-func-config
data:
  name: foo
  newName: bar
  newTag: 4.5.6
`,
			Input: `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: v1
  kind: Pod
  metadata:
    name: the-pod
    namespace: the-namespace
  spec:
    containers:
    - image: foo:1.2.3
      name: test-container
`,
			ExpectedOutput: `apiVersion: v1
kind: Pod
metadata:
  name: the-pod
  namespace: the-namespace
spec:
  containers:
  - image: bar:4.5.6
    name: test-container
`,
		},
		{
			TestName: "set-image should allow specifying an image digest",
			FunctionConfig: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-func-config
data:
  name: foo
  newName: bar
  digest: 123456
`,
			Input: `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: v1
  kind: Pod
  metadata:
    name: the-pod
    namespace: the-namespace
  spec:
    containers:
    - image: foo:1.2.3
      name: test-container
`,
			ExpectedOutput: `apiVersion: v1
kind: Pod
metadata:
  name: the-pod
  namespace: the-namespace
spec:
  containers:
  - image: bar@123456
    name: test-container
`,
		},
		{
			TestName: "set-image should accept a SetImage functionConfig",
			FunctionConfig: `
apiVersion: fn.kpt.dev/v1alpha1
kind: SetImage
metadata:
  name: my-func-config
image:
  name: nginx
  newName: apache
  newTag: 2.4.52
additionalImageFields:
  - kind: MyResource
    create: false
    group: dev.example.com
    path: spec/manifest/images[]/image
    version: v1
`,
			Input: `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: dev.example.com/v1
  kind: MyResource
  metadata:
    name: my-resource
    namespace: my-namespace
  spec:
    containers:
    - image: nginx:1.21.4
      name: my-server
    - image: postgres:14.1
      name: my-store
    manifest:
      images:
      - image: nginx:1.21.4
      - image: postgres:14.1
`,
			ExpectedOutput: `apiVersion: dev.example.com/v1
kind: MyResource
metadata:
  name: my-resource
  namespace: my-namespace
spec:
  containers:
  - image: apache:2.4.52
    name: my-server
  - image: postgres:14.1
    name: my-store
  manifest:
    images:
    - image: apache:2.4.52
    - image: postgres:14.1
`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			output := runImageTransformer(t, tc.Input, tc.FunctionConfig)
			assert.Equal(t, tc.ExpectedOutput, output)
		})
	}

}

func TestFunctionConfigErrors(t *testing.T) {
	testCases := []struct {
		TestName       string
		FunctionConfig string
		ExpectedError  string
	}{
		{
			TestName: "set-image should return an error if image name is unset",
			FunctionConfig: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-func-config
data:
  newName: bar
  newTag: v1.0
`,
			ExpectedError: `missing image name`,
		},
		{
			TestName: "set-image should return an error if image newName, newTag, and digest are unset",
			FunctionConfig: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-func-config
data:
  name: foo
`,
			ExpectedError: `missing image newName, newTag, or digest`,
		},
		{
			TestName: "set-image should return an error when both image newTag and digest are set",
			FunctionConfig: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-func-config
data:
  name: foo
  newName: bar
  newTag: v1.0
  digest: 12345
`,
			ExpectedError: `image newTag and digest both set`,
		},
		{
			TestName: "set-image should return an error when data is missing from functionConfig",
			FunctionConfig: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-func-config
`,
			ExpectedError: `missing image name`,
		},
		{
			TestName: "set-image should return an error when an invalid ConfigMap is used as the functionConfig",
			FunctionConfig: `
apiVersion: v1
kind: ConfigMap
data:
  name:
    unexpected: object
`,
			ExpectedError: "SubObject has unmatched field type: `data",
		},
		{
			TestName: "set-image should return an error when an invalid SetImage is used as the functionConfig",
			FunctionConfig: `
apiVersion: fn.kpt.dev/v1alpha1
kind: SetImage
image:
  name:
    unexpected: object
`,
			ExpectedError: "Resource(apiVersion=fn.kpt.dev/v1alpha1, kind=SetImage, Name=) has unmatched field type: `",
		},
	}

	input := `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: v1
  kind: ConfigMap
`

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			_, err := runImageTransformerE(input, tc.FunctionConfig)
			assert.EqualError(t, err, tc.ExpectedError)
		})
	}
}

func TestAnnotationsTransformerResults(t *testing.T) {
	testCases := []struct {
		TestName        string
		FunctionConfig  string
		Input           string
		ExpectedResults string
	}{
		{
			TestName: `record which image fields were updated`,
			FunctionConfig: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-func-config
data:
  name: foo
  newName: bar
  newTag: 4.5.6
`,
			Input: `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: v1
  kind: Pod
  metadata:
    name: the-pod
    namespace: the-namespace
  spec:
    containers:
    - image: foo:1.2.3
      name: test-container
- apiVersion: v1
  kind: Pod
  metadata:
    name: the-pod2
    namespace: the-namespace
  spec:
    containers:
    - image: foo:1.2.3
      name: test-container
    - image: foo:latest
      name: test-container2
`,
			ExpectedResults: `
[info]: set image from foo:1.2.3 to bar:4.5.6
[info]: set image from foo:1.2.3 to bar:4.5.6
[info]: set image from foo:latest to bar:4.5.6
`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			rl, err := runImageTransformerResults(tc.Input, tc.FunctionConfig)
			assert.Equal(t, nil, err)
			rl.Results.Sort()
			resultStr := "\n"
			for _, r := range rl.Results {
				resultStr += fmt.Sprintf("%s\n", r.String())
			}
			assert.Equal(t, tc.ExpectedResults, resultStr)
		})
	}
}

package main

import (
	"testing"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestStarlarkFunctionConfig(t *testing.T) {
	testcases := []struct {
		config       string
		expectErrMsg string
	}{
		{
			config: `apiVersion: kpt.dev/v1beta1
kind: StarlarkFunction
metadata:
  name: my-star-fn
  namespace: foo
  labels:
    foo: bar
source:
  path: path/to/starlark/script
`,
		},
		{
			config: `apiVersion: kpt.dev/v1beta1
kind: StarlarkFunction
source:
  path: path/to/starlark/script
`,
			expectErrMsg: "name is required in starlark function config",
		},
		{
			config: `apiVersion: kpt.dev/v1beta1
kind: StarlarkFunction
metadata:
  name: my-star-fn
source:
  path: path/to/starlark/script
  url: https://example.com/foo.star
`,
			expectErrMsg: "only one of inline, path and url can be set",
		},
		{
			config: `apiVersion: kpt.dev/v1beta1
kind: StarlarkFunction
metadata:
  name: my-star-fn
source:
  inline: |
    for resource in ctx.resource_list["items"]:
      resource["metadata"]["namespace"] = "helloworld"
  url: https://example.com/foo.star`,
			expectErrMsg: "only one of inline, path and url can be set",
		},
		{
			config: `apiVersion: kpt.dev/v1beta1
kind: StarlarkFunction
metadata:
  name: my-star-fn
source:
  inline: |
    for resource in ctx.resource_list["items"]:
      resource["metadata"]["namespace"] = "helloworld"
  path: path/to/starlark/script
`,
			expectErrMsg: "only one of inline, path and url can be set",
		},
	}
	for _, tc := range testcases {
		var sf StarlarkFunction
		if err := yaml.Unmarshal([]byte(tc.config), &sf); err != nil {
			t.Errorf("unexpcted error: %v", err)
			continue
		}
		err := sf.Validate()
		switch {
		case err != nil && tc.expectErrMsg == "":
			t.Errorf("unexpected error: %v", err)
		case err == nil && tc.expectErrMsg != "":
			t.Errorf("expect error: %v, but got nothing", tc.expectErrMsg)
		case err != nil && err.Error() != tc.expectErrMsg:
			t.Errorf("expect error: %v, but got: %v", tc.expectErrMsg, err)
		}
	}
}

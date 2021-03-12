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
			config: `apiVersion: fn.kpt.dev/v1beta1
kind: StarlarkFunction
metadata:
  name: my-star-fn
  namespace: foo
source: |
  def run(r, ns_value):
    for resource in r:
      resource["metadata"]["namespace"] = ns_value
  run(ctx.resource_list["items"], "baz")
`,
		},
		{
			config: `apiVersion: fn.kpt.dev/v1beta1
kind: StarlarkFunction
source: |
  def run(r, ns_value):
    for resource in r:
      resource["metadata"]["namespace"] = ns_value
  run(ctx.resource_list["items"], "baz")
`,
			expectErrMsg: "`metadata.name` must be set in starlark function config",
		},
		{
			config: `apiVersion: fn.kpt.dev/v1beta1
kind: StarlarkFunction
metadata:
  name: my-star-fn
`,
			expectErrMsg: "`source` must not be empty",
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

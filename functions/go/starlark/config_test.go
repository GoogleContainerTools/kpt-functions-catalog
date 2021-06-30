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
			config: `apiVersion: fn.kpt.dev/v1alpha1
kind: StarlarkRun
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
			config: `apiVersion: fn.kpt.dev/v1alpha1
kind: StarlarkRun
source: |
  def run(r, ns_value):
    for resource in r:
      resource["metadata"]["namespace"] = ns_value
  run(ctx.resource_list["items"], "baz")
`,
			expectErrMsg: "`metadata.name` must be set in the starlark `functionConfig`",
		},
		{
			config: `apiVersion: fn.kpt.dev/v1alpha1
kind: StarlarkRun
metadata:
  name: my-star-fn
`,
			expectErrMsg: "`source` in `StarlarkRun` must not be empty",
		},
		{
			config: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-star-fn
data:
  source: |
    def run(r, ns_value):
      for resource in r:
        resource["metadata"]["namespace"] = ns_value
    run(ctx.resource_list["items"], "baz")
`,
		},
		{
			config: `apiVersion: v1
kind: ConfigMap
data:
  source: |
    def run(r, ns_value):
      for resource in r:
        resource["metadata"]["namespace"] = ns_value
    run(ctx.resource_list["items"], "baz")
`,
			expectErrMsg: "`metadata.name` must be set in the starlark `functionConfig`",
		},
		{
			config: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-star-fn
`,
			expectErrMsg: "`data.source` must not be empty in `ConfigMap`",
		},
		{
			config: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-star-fn
data:
  param1: foo
`,
			expectErrMsg: "`data.source` must not be empty in `ConfigMap`",
		},
	}
	for _, tc := range testcases {
		var sfc StarlarkFnConfig
		if err := yaml.Unmarshal([]byte(tc.config), &sfc); err != nil {
			t.Errorf("unexpcted error: %v", err)
			continue
		}
		err := sfc.Validate()
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

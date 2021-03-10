package main

import (
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
)

const (
	helpText = `Run a starlark script to update or validate resources.

You can specify your starlark script inline under field source like this:

apiVersion: fn.kpt.dev/v1beta1
kind: StarlarkFunction
metadata:
  name: my-star-fn
source: |
  # set the namespace on each resource
  def run(r, ns_value):
    for resource in r:
	  # mutate the resource
	  resource["metadata"]["namespace"] = ns_value
  # get the value to add
  ns_value = ctx.resource_list["functionConfig"]["data"]["foo"]
  run(ctx.resource_list["items"], ns_value)
data:
  foo: baz

You can specify optional key-value pairs in 'data' field and then reference them
in the starlark script like above.
`
)

func main() {
	sf := &StarlarkFunction{}
	resourceList := &framework.ResourceList{
		FunctionConfig: sf,
	}

	cmd := framework.Command(resourceList, func() error {
		err := func() error {
			if ve := sf.Validate(); ve != nil {
				return ve
			}
			if te := sf.Transform(resourceList); te != nil {
				return te
			}
			return nil
		}()
		if err != nil {
			resourceList.Result = &framework.Result{
				Name: "starlark",
				Items: []framework.Item{
					{
						Message:  err.Error(),
						Severity: framework.Error,
					},
				},
			}
			resourceList.FunctionConfig = nil
			return resourceList.Result
		}
		return nil
	})
	cmd.Long = helpText
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

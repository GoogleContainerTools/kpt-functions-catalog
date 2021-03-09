package main

import (
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
)

const (
	helpText = `Run a starlark script to update or validate resources.

There are 3 ways to specify your starlark source:
1) Using inline source
2) Using a file path
3) Using a URL

To specify the starlark source inline, you can use the configuration like this:

apiVersion: kpt.dev/v1beta1
kind: StarlarkFunction
metadata:
  name: my-star-fn
source:
  inline: |
    # set the namespace on each resource
    def run(r, ns_value):
      for resource in r:
        # mutate the resource
        resource["metadata"]["namespace"] = ns_value
    # get the value to add
    ns_value = ctx.resource_list["functionConfig"]["keyValues"]["foo"]
    run(ctx.resource_list["items"], ns_value)
keyValues:
  foo: baz

You can specify optional key-value pairs in 'keyValues' field and then reference
them in the starlark script like above. You can do the same when using path or
URL.

To specify the starlark source using path, you can use the configuration like
the following. You need to mount the source file in the container.

apiVersion: kpt.dev/v1beta1
kind: StarlarkFunction
metadata:
  name: my-star-fn
source:
  path: /path/to/your/starlark/source
keyValues:
  foo: baz

To specify the starlark source using URL, you can use the configuration like the
following. You need to grant network access permission.

apiVersion: kpt.dev/v1beta1
kind: StarlarkFunction
metadata:
  name: my-star-fn
source:
  url: https://raw.githubusercontent.com/mengqiy/starlark-function-test/main/helloworld.star
keyValues:
  foo: baz
`
)

func main() {
	sf := &StarlarkFunction{}
	resourceList := &framework.ResourceList{
		FunctionConfig: sf,
	}

	cmd := framework.Command(resourceList, func() error {
		if err := sf.Transform(resourceList); err != nil {
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

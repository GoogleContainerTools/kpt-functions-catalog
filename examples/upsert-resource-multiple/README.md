# upsert-resource: Multiple Resources Example

In this example, we will see how to upsert multiple resources using `upsert-resource`
function.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/upsert-resource-multiple
```

kpt CLI accepts only one resource as fn-config. Hence, `upsert-resource` function 
accepts Custom Resource `UpsertResourceList` which is used to wrap multiple resources to upsert. 
You can find an example of Custom Resource `UpsertResourceList` at `.fn-config/fn-config.yaml`

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn eval -i upsert-resource:unstable --fn-config .fn-config/fn-config.yaml
```

### Expected result

- Check the resource with name `myService` is replaced with input resource. The
value of field `app` is updated.
- Check that a new resource with name `myDeployment2` is created.

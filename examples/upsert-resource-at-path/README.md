# upsert-resource: Upsert At Specific Path Example

In this example, we will see how to add a resource at specified path using `upsert-resource` function.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/upsert-resource-at-path
```

The input resource is present at path `.fnconfig/fn-config.yaml`. It has an annotation
`config.kubernetes.io/upsert-path` which is used to specify the target path where the resource
should be upserted.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn eval -i upsert-resource:unstable --fn-config .fn-config/fn-config.yaml
```

### Expected result

Check the resource with name `myService` is created in the file at path `subpkg/service.yaml` 

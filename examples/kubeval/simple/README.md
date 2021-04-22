# kubeval: simple example

The `kubeval` KRM config function validates Kubernetes resources using kubeval.
Learn more on the [kubeval website].

This example invokes the kubeval function against the builtin Kubernetes
v1.19.8 schema.

## Function invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/kubeval .
kpt fn run kubeval
```

## Expected Results

This should give the following output:

```sh
[ERROR] Invalid type. Expected: [integer,null], given: string in object 'v1/ReplicationController//bob' in file resources.yaml in field spec.replicas
error: exit status 1
```

In the `resources.yaml` file, replace the value of `spec.replicas`
with an integer to pass validation and rerun the command. This will return
success (no output).

[kubeval website]: https://www.kubeval.com/

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
[ERROR] Additional property templates is not allowed in object 'v1/ReplicationController//bob' in file resources.yaml in field templates
[ERROR] Invalid type. Expected: [integer,null], given: string in object 'v1/ReplicationController//bob' in file resources.yaml in field spec.replicas
error: exit status 1
```

There are validation error in the `resources.yaml` file, to fix them:
- replace the value of `spec.replicas` with an integer
- change `templates` to `template`

Rerun the command, and it will return success (no output).

[kubeval website]: https://www.kubeval.com/

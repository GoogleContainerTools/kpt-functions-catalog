# Set Namespace

The `set-namespace` KRM config function adds or replaces the
`metadata.namespace` field on all resources except for [those known to be
cluster-scoped].

## Function invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-namespace .
kpt fn run set-namespace/configs --fn-path set-namespace/functions
```

## Expected result

Check all resources have `metadata.namespace` set to `example-ns`:

```sh
kpt cfg cat set-namespace/configs
```

[those known to be cluster-scoped]:
  https://github.com/kubernetes-sigs/kustomize/blob/007a5327d7b553d9a8451749fb8b6c9d1de3e482/kyaml/yaml/types.go#L119

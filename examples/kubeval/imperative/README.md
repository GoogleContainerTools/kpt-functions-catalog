# kubeval: Imperative Example

### Overview

This example demonstrates how to imperatively invoke [`kubeval`] function to
validate KRM resources.

### Function invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/kubeval/imperative
kpt fn eval imperative --image=gcr.io/kpt-fn/kubeval:unstable -- strict=true skip_kinds=MyCustom,MyOtherCustom
```

The key-value pair(s) provided after `--` will be converted to ConfigMap by kpt
and used as the function configuration.

We set `strict=true` to disallow unknown fields, and we set
`skip_kinds=MyCustom,MyOtherCustom` to skip 2 kinds that we don't have schemas.

### Expected Results

This should give the following output:

```sh
  Stderr:
    "[ERROR] Additional property templates is not allowed in object 'v1/ReplicationController//bob' in file resources.yaml in field templates"
    "[ERROR] Invalid type. Expected: [integer,null], given: string in object 'v1/ReplicationController//bob' in file resources.yaml in field spec.replicas"
    ""
  Exit Code: 1
```

You should see 2 errors. One is complaining about `templates` is unknown. The
other is about `spec.replicas` is not valid.

To fix them:
- replace the value of `spec.replicas` with an integer
- change `templates` to `template`

Rerun the command, and it should succeed.

[`kubeval`]: https://catalog.kpt.dev/kubeval/v0.1/

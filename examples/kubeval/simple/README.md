# kubeval: Simple Example

### Overview

This example demonstrates how to declaratively run [`kubeval`] function to
validate KRM resources.

The following is the `Kptfile` in this example: 

```yaml
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: example
pipeline:
  validators:
  - image: gcr.io/kpt-fn/kubeval:unstable
    configMap:
      strict: 'true'
```

The function configuration is provided using a ConfigMap. We set 1 key-value
pair:
- `strict: 'true'`: We disallow unknown fields.

### Function invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/kubeval/simple .
kpt fn render simple
```

### Expected Results

This should give the following output:

```sh
Package "simple": 

[RUNNING] "gcr.io/kpt-fn/kubeval:unstable"
[FAIL] "gcr.io/kpt-fn/kubeval:unstable"
  Stderr:
    "[ERROR] Additional property templates is not allowed in object 'v1/ReplicationController//bob' in file resources.yaml in field templates"
    "[ERROR] Invalid type. Expected: [integer,null], given: string in object 'v1/ReplicationController//bob' in file resources.yaml in field spec.replicas"
    ""
  Exit Code: 1
```

There are validation error in the `resources.yaml` file, to fix them:
- replace the value of `spec.replicas` with an integer
- change `templates` to `template`

Rerun the command, and it should output like the following:
```sh
Package "simple": 

[RUNNING] "gcr.io/kpt-fn/kubeval:unstable"
[PASS] "gcr.io/kpt-fn/kubeval:unstable"

Successfully executed 1 function(s) in 1 package(s).
```

[`kubeval`]: https://catalog.kpt.dev/kubeval/v0.1/

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

The function configuration is provided using a `ConfigMap`. We set 2 key-value
pairs:
- `strict: 'true'`: We disallow unknown fields.
- `skip_kinds: MyCustom`: We skip resources of kind `MyCustom`.

### Function invocation

Get this example and try it out by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/kubeval/simple .
$ kpt fn render simple --results-dir=/tmp
```

### Expected Results

Let's take a look at the structured results in `/tmp/results.yaml`:

```yaml
apiVersion: kpt.dev/v1alpha2
kind: FunctionResultList
metadata:
  name: fnresults
exitCode: 1
items:
  - image: gcr.io/kpt-fn/kubeval:unstable
    exitCode: 1
    results:
      - message: Additional property templates is not allowed
        severity: error
        resourceRef:
          apiVersion: v1
          kind: ReplicationController
          name: bob
        field:
          path: templates
        file:
          path: resources.yaml
      - message: 'Invalid type. Expected: [integer,null], given: string'
        severity: error
        resourceRef:
          apiVersion: v1
          kind: ReplicationController
          name: bob
        field:
          path: spec.replicas
        file:
          path: resources.yaml
```

There are validation error in the `resources.yaml` file, to fix them:
- replace the value of `spec.replicas` with an integer
- change `templates` to `template`

Rerun the command, and it should succeed now.

[`kubeval`]: https://catalog.kpt.dev/kubeval/v0.1/

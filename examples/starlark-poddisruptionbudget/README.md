# starlark: PodDisruptionBudget

### Overview

In this example, we are going to demonstrate how to declaratively run the
[`starlark`] function with an inline starlark script as function configuration
to create a `PodDisruptionBudget` object for the `Deployment`.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/starlark-poddisruptionbudget@starlark/v0.1.2
```

We are going to use the following `Kptfile` and `fn-config.yaml` to configure
the function:

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/starlark:v0.1.2
      configPath: fn-config.yaml
```

```yaml
# fn-config.yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: StarlarkRun
metadata:
  name: pdb-for-depl
source: |
  pdb = {
    "apiVersion": "policy/v1beta1",
    "kind": "PodDisruptionBudget",
    "metadata": {
      "name": "nginx-pdb",
    },
    "spec": {
      "minAvailable": 1,
      "selector": {
        "matchLabels": {
          "app": "nginx",
        },
      },
    },
  }
  def is_pdb(r):
    return r["apiVersion"] == "policy/v1beta1" and r["kind"] == "PodDisruptionBudget" and r["metadata"]["name"] == "nginx-pdb"
  def ensure_pdb(resources):
    for resource in resources:
      if is_pdb(resource):
        return
    resources.append(pdb)
  ensure_pdb(ctx.resource_list["items"])
```

The Starlark script is embedded in the `source` field. This script reads the
input KRM resources from `ctx.resource_list` and ensures there is a
`PodDisruptionBudget` object for the nginx `Deployment`.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render starlark-poddisruptionbudget
```

### Expected result

A new file should have been created, and it should contain a
`PodDisruptionBudget` object.

[`starlark`]: https://catalog.kpt.dev/starlark/v0.1/

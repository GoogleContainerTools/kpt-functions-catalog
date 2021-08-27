# remove-annotated-resources

## Overview

<!--mdtogo:Short-->

Removes resources with the annotation `config.kubernetes.io/local-config: true` from the resource list

<!--mdtogo-->

This function provides a quick way for users to prune resources that they may have been working on locally
and don't intend for Config Sync to pick up for validation and hydration

<!--mdtogo:Long-->

## Usage

`remove-annotated-resources` function can be executed imperatively as follows:

```shell
$ kpt fn eval -i remove-annotated-resources:unstable
```

1. Searched for defined resources in a package
2. Deletes the resources with the following annotation:
   `config.kubernetes.io/local-config: true`

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

Consider the following package:

```
├── ...
├── sample
│   ├── applied.yaml
│   ├── local.yaml
```

```yaml
# applied.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: demo-applied
data:
  foo: bar
```

```yaml
# local.yaml
apiVersion: blueprints.cloud.google.com/v1alpha3
kind: ResourceHierarchy
metadata:
  name: sample-hierarchy
  annotations:
    config.kubernetes.io/local-config: "true"
spec:
  config:
    - simple
  parentRef:
    external: "123456789012"
```

Invoke the function in the package directory:

```shell
$ kpt fn eval -i remove-annotated-resources:unstable
```

The resulting package structure would look like this:

```
├── ...
├── sample
│   ├── applied.yaml
```

<!--mdtogo-->

[kpt doc style guide]: https://github.com/GoogleContainerTools/kpt/blob/main/docs/style-guides/docs.md
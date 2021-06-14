# set-labels

### Overview

<!--mdtogo:Short-->

The set-labels function adds a list of labels to all resources. By default, the
function will not only set the labels in `metadata.labels` but also a bunch of
different places where have references to the labels. These settings are
defined [here][commonlabels].

<!--mdtogo-->

### FunctionConfig

<!--mdtogo:Long-->

There are 2 kinds of `functionConfig` supported by this function:

- `ConfigMap`
- A custom resource of kind `SetLabels`

To use a `ConfigMap` as the `functionConfig`, the desired labels must be
specified in the `data` field.

To add 2 labels `color: orange` and `fruit: apple` to all resources, we use the
following `functionConfig`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  color: orange
  fruit: apple
```

To use a `SetLabels` custom resource as the `functionConfig`, the desired labels
must be specified in the `labels` field. You can optionally
use `additionalLabelFields` to specify the additional fields you want to update.
It will be used jointly with the [defaults][commonlabels].
`additionalLabelFields` has following fields:

- `group`: Select the resources by API version group. Will select all groups if
  omitted.
- `version`: Select the resources by API version. Will select all versions if
  omitted.
- `kind`: Select the resources by resource kind. Will select all kinds if
  omitted.
- `path`: Specify the path to the field that the value needs to be updated. This
  field is required.
- `create`: If it's set to true, the field specified will be created if it
  doesn't exist. Otherwise, the function will only update the existing field.

To add 2 labels `color: orange` and `fruit: apple` to all built-in resources and
the path `data.selector` in `MyOwnKind` resource, we use the
following `functionConfig`:

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetLabels
metadata:
  name: my-config
labels:
  color: orange
  fruit: apple
additionalLabelFields:
  - path: data/selector
    kind: MyOwnKind
    create: true
```

<!--mdtogo-->

[commonlabels]: https://github.com/kubernetes-sigs/kustomize/blob/master/api/konfig/builtinpluginconsts/commonlabels.go#L6

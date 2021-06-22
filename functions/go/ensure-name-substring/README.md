# ensure-name-substring

### Overview

<!--mdtogo:Short-->

The `ensure-name-substring` function prepends a prefix or appends a suffix to
the KRM resource names.

This function can be useful to ensure all KRM resources share a common naming
convention and avoid naming conflicting:

- Each team or project must use its name as the prefix for their KRM resource
  names.
- The environment name (e.g. one of prod, staging and test) must be used as the
  suffix for it KRM resources.

<!--mdtogo-->

You can learn more about names [here][names].

Note: This is an alpha function, and we are actively seeking feedback on the
`functionConfig` syntax and behavior. If you have any suggestion or feedback,
please file an [issue].

<!--mdtogo:Long-->

### Usage

This function can be used with any KRM function orchestrators (e.g. kpt).

This function is **idempotent**. If the desired name substring is already part
of the name, the function takes no actions. Otherwise, it prepends or appends
the name substring depending on the `functionConfig`.

Note: If the original name of a resource happens to contain the desired
substring, the desired substring will not be added again. Users need to ensure
the name collisions don't happen.

This function does Not process the following resources:

- `CustomResourceDefinition`
- `Namespace`
- `APIService`

In addition to updating the `metadata.name` field for each resource, the
function will also update the [fields][namereference] that references the name
field. e.g. if the name of a `ConfigMap` got updated and this `ConfigMap` is
being referenced in `Volumes` in a `Pod`, field `spec.volumes.configMap.name`
will also be updated.

This function can be used both declaratively and imperatively.

#### FunctionConfig

There are 2 kinds of `functionConfig` supported by this function:

- `ConfigMap`
- A custom resource of kind `EnsureNameSubstring`

To use a `ConfigMap` as the `functionConfig`, only one key-value pair is allowed
in `data` field. The key must be one of `prepend` and `append`, and the value
must be the desired name substring.

For example, to ensure a name substring `dev-` exists in all resource names, we
use the following `functionConfig`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-fn-config
data:
  prepend: dev-
```

To use a `EnsureNameSubstring` custom resource as the `functionConfig`, the
desired substring must be specified in the `substring` field, and you can
specify either `prepend` or `append` in the `editMode` field. If `editMode` is
unspecified, `prepend` will be used.

Sometimes you have resources (especially custom resources) that have name fields
in fields other than `metadata.name`, you can specify such name fields
using `additionalNameFields`. It will be used jointly with
the [defaults][defaultnamefields].

`additionalNameFields` has the following fields:

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

For example, to ensure a name substring `dev-` exists in all built-in resources
and custom resources of kind `MyOwnResource` in field `spec.name`, we use the
following `functionConfig`:

```yaml
apiVersion: v1
kind: EnsureNameSubstring
metadata:
  name: my-fn-config
substring: dev-
editMode: prepend
additionalNameFields:
  - path: spec/name
    kind: MyOwnResource
```

<!--mdtogo-->

[names]: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/

[issue]: https://github.com/GoogleContainerTools/kpt/issues/new/choose

[namereference]: https://github.com/kubernetes-sigs/kustomize/blob/master/api/konfig/builtinpluginconsts/namereference.go#L7

[defaultnamefields]: https://github.com/kubernetes-sigs/kustomize/blob/master/api/konfig/builtinpluginconsts/nameprefix.go#L7

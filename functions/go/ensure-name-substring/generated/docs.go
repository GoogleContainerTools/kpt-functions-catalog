// Code generated by "mdtogo"; DO NOT EDIT.
package generated

var EnsureNameSubstringShort = `Ensure a name substring is part of the name.`
var EnsureNameSubstringLong = `
Note: This is an alpha function, and we are actively seeking feedback on the
function config syntax and behavior. If you have suggestion or feedback, please
file an issue [here](https://github.com/GoogleContainerTools/kpt/issues/new/choose).

If the desired name substring is already part of the name, it takes no actions.
Otherwise, it prepends or appends the name substring.

Note: If the original name of a resource happens to contain the desired
substring, the desired substring will not be added again. Users need to ensure
the name collisions don't happen.

To configure it using a ConfigMap, only one key-value pair is allowed in ` + "`" + `data` + "`" + `
field. The key must be one of ` + "`" + `prepend` + "`" + ` and ` + "`" + `append` + "`" + `, and the value is the
desired name substring. 

  prepend|append: Desired name substring

For example: To ensure a name substring ` + "`" + `dev-` + "`" + ` exists in all resources and
prepends it if not found.

  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: my-config
  data:
    prepend: dev-

This function does Not process the following resources:
- CustomResourceDefinition
- Namespace
- APIService

You can optionally use key ` + "`" + `fieldSpecs` + "`" + ` to specify the resource selector you
want to use. By default, the function will not only update the ` + "`" + `metadata/name` + "`" + `
but also a bunch of different places where have references to the names. These
field specs are defined in
https://github.com/kubernetes-sigs/kustomize/blob/master/api/konfig/builtinpluginconsts/namereference.go

To support your own CRDs you will need to add more items to fieldSpecs list.
Your own specs will be used with the default ones.

Field spec has following fields:

- group: Select the resources by API version group. Will select all groups
  if omitted.
- version: Select the resources by API version. Will select all versions
  if omitted.
- kind: Select the resources by resource kind. Will select all kinds
  if omitted.
- path: Specify the path to the field that the value will be updated. This field
  is required.
- create: If it's set to true, the field specified will be created if it doesn't
  exist. Otherwise, the function will only update the existing field.

For more information about fieldSpecs, please see
https://kubectl.docs.kubernetes.io/guides/extending_kustomize/builtins/#arguments-8

For example, to ensure ` + "`" + `dev-` + "`" + ` also exists ` + "`" + `spec/name` + "`" + ` in ` + "`" + `MyOwnResource` + "`" + `:

  apiVersion: v1
  kind: EnsureNameSubstring
  metadata:
    name: my-config
  substring: dev-
  editMode: prepend
  fieldSpecs:
  - path: spec/name
    kind: MyOwnResource
`
var EnsureNameSubstringExamples = `
https://github.com/GoogleContainerTools/kpt-functions-catalog/tree/master/examples/ensure-name-substring/
`

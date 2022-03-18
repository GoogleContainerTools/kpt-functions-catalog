

// Code generated by "mdtogo"; DO NOT EDIT.
package generated

var SetNamespaceShort = `The ` + "`" + `set-namespace` + "`" + ` function replaces the ` + "`" + `namespace` + "`" + ` specific resource type in a variety
of KRM resources.`
var SetNamespaceLong = `
## Usage

This function can be used with any KRM function orchestrators (e.g. kpt).

- If the resource is ` + "`" + `Namespace` + "`" + `, ` + "`" + `set-namespace` + "`" + ` updates the ` + "`" + `metadata.name` + "`" + ` field.
- If the resource is ` + "`" + `RoleBinding` + "`" + ` or ` + "`" + `ClusterRoleBinding` + "`" + ` resource, the function updates 
  the namespace field in the ` + "`" + `subjects` + "`" + ` element whose name is ` + "`" + `default` + "`" + `.
- If the resource is ` + "`" + `CustomResourceDefinition` + "`" + ` (CRD), ` + "`" + `set-namespace` + "`" + ` updates the 
  ` + "`" + `spec/conversion/webhook/clientConfig/service/namespace` + "`" + ` field.
- If the resource is ` + "`" + `APIService` + "`" + `, ` + "`" + `set-namespace` + "`" + ` updates the
  ` + "`" + `spec/service/namespace` + "`" + ` field.
- If there is a [` + "`" + `depends-on` + "`" + `] annotation for a namespaced resource, the namespace
  section of the annotation will be updated if the referenced resource is also
  declared in the package.

  apiVersion: v1
  kind: ServiceAccount
  metadata:
    name: sa
    namespace: example
    annotations:
      config.kubernetes.io/depends-on: /namespaces/example/ServiceAccount/foo # <= this will NOT be updated (resource not declared)
  ---
  kind: RoleBinding
  apiVersion: rbac.authorization.k8s.io/v1
  metadata:
    ...
    annotations:
      config.kubernetes.io/depends-on: /namespaces/example/ServiceAccount/sa # <== this will be updated (resource declared)
  subjects:
    - kind: ServiceAccount
      name: default # <================== name default is used
      namespace: example # <================== this will be updated
  roleRef:
    kind: Role
    name: confluent-operator
    apiGroup: rbac.authorization.k8s.io

This function can be used both declaratively and imperatively.

### FunctionConfig

There are 2 kinds of ` + "`" + `functionConfig` + "`" + ` supported by this function:

- ` + "`" + `ConfigMap` + "`" + `
- A custom resource of kind ` + "`" + `SetNamespace` + "`" + `

To use a ` + "`" + `ConfigMap` + "`" + ` as the ` + "`" + `functionConfig` + "`" + `, the desired namespace must be
specified in the ` + "`" + `data.namespace` + "`" + ` field.

To add a namespace ` + "`" + `staging` + "`" + ` to all resources, we use the
following ` + "`" + `functionConfig` + "`" + `:

  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: my-config
  data:
    namespace: staging
`

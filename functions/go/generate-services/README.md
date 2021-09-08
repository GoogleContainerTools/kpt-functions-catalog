# generate-services

## Overview

<!--mdtogo:Short-->

The generate-services kpt function generates [Config Controller Service resources](https://cloud.google.com/config-connector/docs/reference/resource-docs/serviceusage/service) which are [GCP Services](https://cloud.google.com/service-usage/docs) required for actuating GCP resources.

<!--mdtogo-->

This removes the requirement that users explicitly enable services, but more importantly, it makes it easier to author kpt packages, because Service resources no longer need to be embedded in every package and there's no need for a standard naming scheme across package providers.

Services produced by this function are also de-duplicated, meaning there will only ever be one Service resource enabling that service in that particular project. This means that the `cnrm.cloud.google.com/disable-on-destroy: "true"` annotation is no longer required, and services can now be disabled when they are no longer in use, as long as all the Resources for that project are managed in the same source directory or repository.

## Input

This function accepts all KRM resources, but ignores those that are not managed by Config Controller.

The following config options are supported:
- `namespace` - namespace set on the generated Services (default: `gcp-services`)
- `disable-on-destroy` - value for the `cnrm.cloud.google.com/disable-on-destroy` annotation on the generated Services (default: `""` - no annotation)

These options can be configured declaratively in the data field of a `ConfigMap` or imperatively following a `kpt fn eval` command.

## Output

Executing the generate-services function will generate a unique set of Service resources.

Naming pattern:

```
${PROJECT_ID}-${HOSTNAME_PREFIX}
```

The `HOSTNAME_PREFIX` will be the Service API Hostname without the `.googleapis.com` suffix, with periods replaced by hyphens (ex: `compute.googleapis.com` -> `compute`).

File path pattern:

```
${NAMESPACE}/service_${NAME}.yaml
```

<!--mdtogo:Long-->

## Usage

1. Create any Config Controller resource(s), in one or more yaml files, anywhere in a directory.
2. Create a ConfigMap to pass optional data to the containerized kpt function as follows (example):
    Example:
    ```
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: generate-services
    data:
      namespace: gcp-services
      disable-on-destroy: true
    ```
3. (Optional) Set the `data.namespace` field to specify which namespace Services will be created in.
4. (Optional) Set the `data.disable-on-destroy` field to specify the value of the `cnrm.cloud.google.com/disable-on-destroy` annotation.
5. (Optional) Add the `config.kubernetes.io/local-config: "true"` annotation to tell ConfigSync to exclude the ConfigMap resource when applying.
6. Run [kpt fn render](https://kpt.dev/book/04-using-functions/01-declarative-function-execution) on the directory containing the resource yaml files.
    Example:
    ```
    kpt fn render .
    ```

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

## Input Example

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: generate-services
  annotations:
    config.kubernetes.io/local-config: "true"
data:
  namespace: gcp-services
  disable-on-destroy: "false"
---
apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
kind: Project
metadata:
  name: example-project-id
  namespace: projects
  annotations:
    cnrm.cloud.google.com/organization-id: "123456789012"
spec: {}
```

## Output Example

All of the inputs will exist in the output, as well as any newly added Service resources:

```
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: example-project-id-cloudresourcemanager
  namespace: gcp-services
  annotations:
    cnrm.cloud.google.com/disable-on-destroy: "false"
    cnrm.cloud.google.com/project-id: example-project-id
    config.kubernetes.io/path: 'gcp-services/service_example-project-id-cloudresourcemanager.yaml'
spec:
  resourceID: cloudresourcemanager.googleapis.com
```

<!--mdtogo-->

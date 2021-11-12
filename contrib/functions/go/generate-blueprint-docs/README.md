# generate-blueprint-docs

## Overview

<!--mdtogo:Short-->

The `generate-blueprint-docs` function generates opinionated documentation for a [blueprint](https://github.com/GoogleCloudPlatform/blueprints).

It works by analyzing a blueprint kpt package to generate markdown documentation and writes it to a file.

<!--mdtogo-->

<!--mdtogo:Long-->

## Usage

The `generate-blueprint-docs` function is expected to be executed imperatively like:

```shell
kpt fn eval -i gcr.io/kpt-fn-contrib/generate-blueprint-docs:unstable --include-meta-resources \
--mount type=bind,src="$(pwd)",dst="/tmp",rw=true
```

## FunctionConfig

This function supports `ConfigMap` `functionConfig`.

- A `readme-path` can be provided to write to a specific file. If a `readme-path` is not provided, it defaults to `tmp/README.md`.

- A `repo-path` can be provided to include in usage section. This will generate a usage section with sample commands like `kpt pkg get ${repo-path}/{pkgname}@version`. If a `repo-path` is not provided, it defaults to `https://github.com/GoogleCloudPlatform/blueprints.git/catalog`.

The `generate-blueprint-docs` function does the following:

1. Scans the package contents including meta resources like Kptfile(s) and function configs.
1. Generates readme contents in markdown format.
1. Writes the generated readme to `readme-path`.

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

Let's start with a sample blueprint.

```yaml
# Kptfile
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: bucket
  annotations:
    blueprints.cloud.google.com/title: Google Cloud Storage Bucket blueprint
info:
  description: A Google Cloud Storage bucket
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configPath: setters.yaml
---
# bucket.yaml
apiVersion: storage.cnrm.cloud.google.com/v1beta1
kind: StorageBucket
metadata:
  name: blueprints-project-bucket # kpt-set: ${project-id}-${name}
  namespace: config-control # kpt-set: ${namespace}
  annotations:
    cnrm.cloud.google.com/force-destroy: "false"
    cnrm.cloud.google.com/project-id: blueprints-project # kpt-set: ${project-id}
spec:
  location: us-central1
  storageClass: standard # kpt-set: ${storage-class}
  uniformBucketLevelAccess: true
  versioning:
    enabled: false
---
# setters.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: setters
data:
  name: bucket2
  namespace: config-control
  project-id: project-id
  storage-class: standard
```

Invoke the function:

```shell
kpt fn eval -i gcr.io/kpt-fn-contrib/generate-blueprint-docs:unstable --include-meta-resources \
--mount type=bind,src="$(pwd)",dst="/tmp",rw=true -- path=/tmp/README.md
```

The following readme will be created:

# Google Cloud Storage Bucket blueprint

A Google Cloud Storage bucket

## Setters

|     Name      |       Value        | Type | Count |
|---------------|--------------------|------|-------|
| name          | bucket             | str  |     1 |
| namespace     | config-control     | str  |     1 |
| project-id    | blueprints-project | str  |     2 |
| storage-class | standard           | str  |     1 |

## Subpackages

This package has no sub-packages.

## Resources

|    File     |              APIVersion               |     Kind      |           Name            |   Namespace    |
|-------------|---------------------------------------|---------------|---------------------------|----------------|
| bucket.yaml | storage.cnrm.cloud.google.com/v1beta1 | StorageBucket | blueprints-project-bucket | config-control |

## Resource References

- [StorageBucket](https://cloud.google.com/config-connector/docs/reference/resource-docs/storage/storagebucket)

## Usage

1.  Clone the package:
    ```
    kpt pkg get https://github.com/GoogleCloudPlatform/blueprints.git/catalog/bucket@${VERSION}
    ```
    Replace `${VERSION}` with the desired repo branch or tag
    (for example, `main`).

1.  Move into the local package:
    ```
    cd "./bucket/"
    ```

1.  Edit the function config file(s):
    - setters.yaml

1.  Execute the function pipeline
    ```
    kpt fn render
    ```

1.  Initialize the resource inventory
    ```
    kpt live init --namespace ${NAMESPACE}"
    ```
    Replace `${NAMESPACE}` with the namespace in which to manage
    the inventory ResourceGroup (for example, `config-control`).

1.  Apply the package resources to your cluster
    ```
    kpt live apply
    ```

1.  Wait for the resources to be ready
    ```
    kpt live status --output table --poll-until current
    ```

<!--mdtogo-->

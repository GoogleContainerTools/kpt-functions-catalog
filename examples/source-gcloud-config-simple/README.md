# source-gcloud-config: Simple Example

### Overview

This example demonstrates how to imperatively run [`source-gcloud-config`] function
to add a ConfigMap resource containing the gcloud configurations.

### Before you begin

1. Install [gcloud SDK](https://cloud.google.com/sdk/docs/install)
1. [Setting defaults for gcloud commands](https://cloud.google.com/artifact-registry/docs/gcloud-defaults)
   Suggest setting up your `project`, `region` and `zone`

### Fetch the executable

This package contains the `source-gcloud-config` executable.
```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/source-gcloud-config-simple
```

### Run the function

```shell
$ kpt fn eval --exec ./source-gcloud-config .
```

### Expected result

A new file `configmap_gcloud-config.kpt.dev.yaml` should be added which contains a ConfigMap resource
 and have your gcloud default configurations stored in the `.data` field.
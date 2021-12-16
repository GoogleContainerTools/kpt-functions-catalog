# preview-hierarchy

## Overview

<!--mdtogo:Short-->

This functions generates a visual hierarchy of [Folder](https://cloud.google.com/config-connector/docs/reference/resource-docs/resourcemanager/folder)
resources from the supplied KRM resources.

<!--mdtogo-->

The purpose of this function is to generate a resource hierarchy from KCC KRMs
passed to it as input. It serves as visual preview SVG or tree output that can be reviewed
before actuating resources / folder structures within a GCP org.

<!--mdtogo:Long-->

### FunctionConfig

<!--mdtogo:Long-->

ConfigMap can be used configure the `preview-hierarchy` function. The output format
for the hierarchy can be provided as key-value pairs using `data` field. The key-values
can be in two forms:
1. Specifying rendering of the tree as stdout and part of results.yaml. ConfigMap as follows:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: preview-hierarchy-func-config
data:
  renderer: text
```

2. Specifying rendering as an output SVG file based on the gcp-draw API. ConfigMap as follows:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: preview-hierarchy-func-config
data:
  output: test.svg
```

<!--mdtogo-->

## Usage

`preview-hierarchy` function is expected to be executed imperatively as follows:

For SVG output

```shell
$ kpt fn eval -i gcr.io/kpt-fn/preview-hierarchy:v0.1.0 -- output=test.svg
```

To print tree visualization to stdout 

```shell
$ kpt fn eval -i gcr.io/kpt-fn/preview-hierarchy:v0.1.0 -- renderer=text
```

`preview-hierarchy` function performs the following steps:

1. Iterates through all [Folder](https://cloud.google.com/config-connector/docs/reference/resource-docs/resourcemanager/folder) resources provided in the package
2. Generates and renders a visual hierarchy of the resources as stdout or as an SVG file depening on the output parameter provided

<!--mdtogo-->

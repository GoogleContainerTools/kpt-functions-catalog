# inflate-helm-chart

[helm]: https://helm.sh/
[charts]: https://helm.sh/docs/topics/charts/
[local]: https://github.com/GoogleContainerTools/kpt-functions-catalog/tree/master/examples/inflate-helm-chart-local
[remote]: https://github.com/GoogleContainerTools/kpt-functions-catalog/tree/master/examples/inflate-helm-chart-remote


## Overview

<!--mdtogo:Short-->

The `inflate-helm-chart` function inflates a local or remote Helm chart. 

<!--mdtogo-->

Helm is a package manager for kubernetes that uses a packaging format
called charts. A chart is a collection of files within a directory, which 
contain templates, CRDs, values, and metadata. 

This function renders charts locally by using the 
[helm template command][https://helm.sh/docs/helm/helm_template/],
so that helm charts can be inflated without needing to install the
helm binary directly.

You can learn more about helm [here][helm] and more about helm
charts [here][charts].

<!--mdtogo:Long-->

## Usage

This function can be used with any KRM function orchestrators such as kpt
or kustomize to inflate a specified helm chart.

In kpt, the function can only be run imperatively. The function either
needs network access to inflate a remote chart or needs a local file to be mounted
into the container to inflate a local chart. As a result, to run the
function with `kpt fn eval`, the flag `--network` must be used for remote charts,
and the flag `--mount` must be used for local charts. See the examples for inflating
[local] and [remote] charts.

It can be used declaratively when run with kustomize. To run the function with kustomize,
the `network` field is needed for remote charts and the `mounts` field is needed for local charts.

### FunctionConfig

<!--mdtogo:Long-->

There are 2 kinds of `functionConfig` supported by this function:

- `ConfigMap`
- A custom resource of kind `InflateHelmChart`

#### `ConfigMap`
To use a `ConfigMap` as the `functionConfig`, the desired parameters must be
specified in the `data` field:

```yaml
data:
  chartHome: string
  configHome: string
  name: string
  version: string
  repo: string
  releaseName: string
  namespace: string
  valuesFile: string
  includeCRDs: string
```

| Field        |  Description | 
| -----------: |  ----------- | 
`helmGlobals`  | Parameters applied to all Helm charts
`helmCharts`   | An array of helm chart parameters
`chartHome`    | A filepath to a directory of charts. The function will look for the chart in this local directory before attempting to pull the chart from a specified repo. Defaults to "tmp/charts". When run in a container, this path MUST have the prefix "tmp/".
`configHome`   | (exec only) Defines a value that the function should pass to helm via the HELM_CONFIG_HOME environment variable. If omitted, {tmpDir}/helm is used, where {tmpDir} is some temporary directory created by the function for the benefit of helm.
`name`         | The name of the chart
`version`      | The version of the chart
`repo`         | A URL locating the chart on the internet
`releaseName`  | Replaces RELEASE_NAME in the chart template output
`namespace`    | Sets the target namespace for a release (.Release.Namespace in the template)
`valuesFile`   | valuesFile is a remote or local file path to a values file to use _instead of_ the default values that accompanied the chart. The default values are in '{ChartHome}/{Name}/values.yaml'.
`includeCRDs`  | Specifies if Helm should also generate CustomResourceDefinitions. Can be 'true' or 'false', defaults to 'false'

The only required field is `name`.

#### `InflateHelmChart`
A `functionConfig` of kind `InflateHelmChart` has the following supported parameters: 

```yaml
helmGlobals:
  chartHome: string
  configHome: string
helmCharts:
- name: string
  version: string
  repo: string
  releaseName: string
  namespace: string
  valuesInline: map[string]interface{}
  valuesFile: string
  valuesMerge: string
  includeCRDs: bool
```

| Field        |  Description | 
| -----------: |  ----------- | 
`helmGlobals`  | Parameters applied to all Helm charts
`helmCharts`   | An array of helm chart parameters
`chartHome`    | A filepath to a directory of charts. The function will look for the chart in this local directory before attempting to pull the chart from a specified repo. Defaults to "tmp/charts". When run in a container, this path MUST have the prefix "tmp/".
`configHome`   | (exec only) Defines a value that the function should pass to helm via the HELM_CONFIG_HOME environment variable. If omitted, {tmpDir}/helm is used, where {tmpDir} is some temporary directory created by the function for the benefit of helm.
`name`         | The name of the chart
`version`      | The version of the chart
`repo`         | A URL locating the chart on the internet
`releaseName`  | Replaces RELEASE_NAME in the chart template output
`namespace`    | Sets the target namespace for a release (.Release.Namespace in the template)
`valuesInline` | Values to use instead of default values that accompany the chart
`valuesFile`   | valuesFile is a remote or local file path to a values file to use _instead of_ the default values that accompanied the chart. The default values are in '{ChartHome}/{Name}/values.yaml'.
`valuesMerge`  | ValuesMerge specifies how to treat ValuesInline with respect to Values. Legal values: 'merge', 'override' (default), 'replace'.
`includeCRDs`  | Specifies if Helm should also generate CustomResourceDefinitions, defaults to 'false'

The only required field is `name`.

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

### Example with kpt

To inflate a remote minecraft chart, you can run the following command: 

```shell
$ kpt fn eval --image gcr.io/kpt-fn/inflate-helm-chart:unstable --network -- \
name=minecraft \
repo=https://itzg.github.io/minecraft-server-charts \
releaseName=test
```

This will add two files to your directory, which you can view:

```shell
$ kpt pkg tree
├── [secret_test-minecraft.yaml]  Secret test-minecraft
└── [service_test-minecraft.yaml]  Service test-minecraft
```

### Example with kustomize
You can specify your `functionConfig` via the `generators` field in your `kustomization.yaml` file:

The `functionConfig` can be of type `InflateHelmChart`:

```yaml
generators:
- |-
  apiVersion: v1
  kind: InflateHelmChart
  metadata:
    name: demo
    annotations:
      config.kubernetes.io/function: |
        container:
          network: true
          image: gcr.io/kpt-fn/inflate-helm-chart:latest
  helmCharts:
  - name: minecraft
    repo: https://itzg.github.io/minecraft-server-charts
    version: 3.1.3
    releaseName: test
```

You can equivalently use a `functionConfig` of type `ConfigMap`:

```yaml
generators:
- |-
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: demo
    annotations:
      config.kubernetes.io/function: |
        container:
          network: true
          image: gcr.io/kpt-fn/inflate-helm-chart:unstable
  data:
    name: minecraft
    repo: https://itzg.github.io/minecraft-server-charts
    version: 3.1.3
    releaseName: test
```

For both of the above kustomizations, the helm chart can be
inflated with the following command:

```shell
kustomize build --enable-alpha-plugins --network .
```

This gives the output:
```yaml
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: test-minecraft
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: test
  name: test-minecraft
spec:
  ports:
  - name: minecraft
    port: 25565
    protocol: TCP
    targetPort: minecraft
  selector:
    app: test-minecraft
  type: ClusterIP
```

<!--mdtogo-->

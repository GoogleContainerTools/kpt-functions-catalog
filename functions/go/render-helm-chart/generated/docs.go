

// Code generated by "mdtogo"; DO NOT EDIT.
package generated

var RenderHelmChartShort = `The ` + "`" + `render-helm-chart` + "`" + ` function renders a local or remote Helm chart.`
var RenderHelmChartLong = `
## Usage

This function can be used with any KRM function orchestrators such as kpt
or kustomize to render a specified helm chart.

In kpt, the function can only be run imperatively. The function either
needs network access to render a remote chart or needs a local file to be mounted
into the container to render a local chart. As a result, to run the
function with ` + "`" + `kpt fn eval` + "`" + `, the flag ` + "`" + `--network` + "`" + ` must be used for remote charts,
and the flag ` + "`" + `--mount` + "`" + ` must be used for local charts. See the examples for inflating
[local] and [remote] charts.

It can be used declaratively when run with kustomize. To run the function with kustomize,
the ` + "`" + `network` + "`" + ` field is needed for remote charts and the ` + "`" + `mounts` + "`" + ` field is needed for local charts.

### FunctionConfig


There are 2 kinds of ` + "`" + `functionConfig` + "`" + ` supported by this function:

- ` + "`" + `ConfigMap` + "`" + `
- A custom resource of kind ` + "`" + `RenderHelmChart` + "`" + `

` + "`" + `ConfigMap` + "`" + `:
To use a ` + "`" + `ConfigMap` + "`" + ` as the ` + "`" + `functionConfig` + "`" + `, the desired parameters must be
specified in the ` + "`" + `data` + "`" + ` field:

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


| Field        |  Description | Example
| -----------: |  ----------- | -----------
` + "`" + `chartHome` + "`" + `    | A filepath to a directory of charts. The function will look for the chart in this local directory before attempting to pull the chart from a specified repo. Defaults to "tmp/charts". When run in a container, this path MUST have the prefix "tmp/". | tmp/charts
` + "`" + `configHome` + "`" + `   | Defines a value that the function should pass to helm via the HELM_CONFIG_HOME environment variable. If omitted, {tmpDir}/helm is used, where {tmpDir} is some temporary directory created by the function for the benefit of helm. This option is not supported when running in a container. It is supported only in exec mode (e.g. with kustomize) | /tmp/helm/config
` + "`" + `name` + "`" + `         | The name of the chart | minecraft
` + "`" + `version` + "`" + `      | The version of the chart | 3.1.3
` + "`" + `repo` + "`" + `         | A URL locating the chart on the internet | https://itzg.github.io/minecraft-server-charts
` + "`" + `releaseName` + "`" + `  | Replaces RELEASE_NAME in the chart template output | test
` + "`" + `namespace` + "`" + `    | Sets the target namespace for a release (` + "`" + `.Release.Namespace` + "`" + ` in the template) | my-namespace
` + "`" + `valuesFile` + "`" + `   | valuesFile is a remote or local file path to a values file to use instead of the default values that accompanied the chart. The default values are in '{chartHome}/{name}/values.yaml', where ` + "`" + `chartHome` + "`" + ` and ` + "`" + `name` + "`" + ` are the parameters defined above. | Using a local values file: path/to/your/values.yaml <br> <br> Using a remote values file: https://raw.githubusercontent.com/config-sync-examples/helm-components/main/cert-manager-values.yaml
` + "`" + `includeCRDs` + "`" + `  | Specifies if Helm should also generate CustomResourceDefinitions. Legal values: "true", "false" (default). | "true"


The only required field is ` + "`" + `name` + "`" + `.

` + "`" + `RenderHelmChart` + "`" + `:
A ` + "`" + `functionConfig` + "`" + ` of kind ` + "`" + `RenderHelmChart` + "`" + ` has the following supported parameters: 

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

| Field        |  Description | Example
| -----------: |  ----------- | -----------
` + "`" + `helmGlobals` + "`" + `  | Parameters applied to all Helm charts
` + "`" + `helmCharts` + "`" + `   | An array of helm chart parameters
` + "`" + `chartHome` + "`" + `    | A filepath to a directory of charts. The function will look for the chart in this local directory before attempting to pull the chart from a specified repo. Defaults to "tmp/charts". When run in a container, this path MUST have the prefix "tmp/". | tmp/charts
` + "`" + `configHome` + "`" + `   | Defines a value that the function should pass to helm via the HELM_CONFIG_HOME environment variable. If omitted, {tmpDir}/helm is used, where {tmpDir} is some temporary directory created by the function for the benefit of helm. This option is not supported when running in a container. It is supported only in exec mode (e.g. with kustomize) | /tmp/helm/config
` + "`" + `name` + "`" + `         | The name of the chart | minecraft
` + "`" + `version` + "`" + `      | The version of the chart | 3.1.3
` + "`" + `repo` + "`" + `         | A URL locating the chart on the internet | https://itzg.github.io/minecraft-server-charts
` + "`" + `releaseName` + "`" + `  | Replaces RELEASE_NAME in the chart template output | test
` + "`" + `namespace` + "`" + `    | Sets the target namespace for a release (` + "`" + `.Release.Namespace` + "`" + ` in the template) | my-namespace
` + "`" + `valuesInline` + "`" + ` | Values to use instead of default values that accompany the chart |  global: <br> &emsp; enabled: false <br> tests: <br> &emsp; enabled: false  
` + "`" + `valuesFile` + "`" + `   | valuesFile is a remote or local file path to a values file to use instead of the default values that accompanied the chart. The default values are in '{chartHome}/{name}/values.yaml', where ` + "`" + `chartHome` + "`" + ` and ` + "`" + `name` + "`" + ` are the parameters defined above. | Using a local values file: path/to/your/values.yaml <br> <br> Using a remote values file: https://raw.githubusercontent.com/config-sync-examples/helm-components/main/cert-manager-values.yaml
` + "`" + `valuesMerge` + "`" + `  | ValuesMerge specifies how to treat ValuesInline with respect to Values. Legal values: 'merge', 'override' (default), 'replace'. | replace
` + "`" + `includeCRDs` + "`" + `  | Specifies if Helm should also generate CustomResourceDefinitions. Defaults to false. | true

The only required field is ` + "`" + `name` + "`" + `.
`
var RenderHelmChartExamples = `
To render a remote minecraft chart, you can run the following command: 

  $ kpt fn eval --image gcr.io/kpt-fn/render-helm-chart:unstable --network -- \
  name=minecraft \
  repo=https://itzg.github.io/minecraft-server-charts \
  releaseName=test

The key-value pairs after the ` + "`" + `--` + "`" + ` will be converted to a ` + "`" + `functionConfig` + "`" + ` of kind
` + "`" + `ConfigMap` + "`" + `. The above command will add two files to your directory, which you can view:

  $ kpt pkg tree
  ├── [secret_test-minecraft.yaml]  Secret test-minecraft
  └── [service_test-minecraft.yaml]  Service test-minecraft
`

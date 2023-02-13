// Code generated by "mdtogo"; DO NOT EDIT.
package generated

var SourceGcloudConfigShort = `This ` + "`" + `source-gcloud-config` + "`" + ` function adds a ConfigMap resource which contains the gcloud configurations.`
var SourceGcloudConfigLong = `
## Usage

This function can be used as [variant constructor](https://github.com/GoogleContainerTools/kpt/issues/2590)
which works with ` + "`" + `kpt pkg get` + "`" + ` hook to add gcloud configurations for deployable package instance.

This function follows a convention to map gcloud configs to KRM resource configurations and store them in a ConfigMap 
resource. This ConfigMap later on can be used by other kpt functions as FunctionConfig. This simplifies the user experience
on configuring the custom functionConfig for their kpt packages. 

Since this ConfigMap is not expected to be deployed to a cluster, it is applied to the ` + "`" + `local-config` + "`" + ` annotation.   
`
var SourceGcloudConfigExamples = `
Commands
` + "`" + `go build -o source-gcloud-config` + "`" + `
` + "`" + `kpt fn eval --exec ./source-gcloud-config` + "`" + `

Output:
A new ` + "`" + `ConfigMap_gcloud-config.kpt.dev.yaml` + "`" + ` file is added to your pkg directory. 

  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: gcloud-config.kpt.dev
    annotations:
      config.kubernetes.io/local-config: "true"
  data:
    domain: <YOUR ACCOUNT DOMAIN>
    namespace: <GCP PROJECT ID>
    orgID: <YOUR ORGANIZATION ID>
    projectID: <GCP PROJECT ID>
    region: <GCLOUD CONFIG REGION>
    zone: <GCLOUD CONFIG ZONE>
`

<!-- Copyright 2020 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. -->

# Helm-template Usage

Kpt packages are just configuration so any solution, like the `helm template` command, which emits configuration can also be used to generate kpt packages. `Helm-template` is a kpt function which generates a new kpt package from a local Helm chart or upserts Helm chart configuration to an existing kpt package. In the context of a pipeline, these packages can then be further customized using other kpt functions.

## FAQs

### How can I set arbitrary values in my chart using `--set`

We recommend that you create a new values.yaml file with the values you want so you can check the new file into a version-controlled repository. You can specify an optional `values_path` argument to the helm-template command containing the relative path to your new file.  
`$ docker run -v $(pwd)/charts/bitnami:/source gcr.io/kpt-functions/helm-template chart_path=/source/redis name=my-redis` **`values_path=/source/redis/values-production.yaml`**

## Examples

### Example 1: Expand and apply multiple charts to a cluster

#### Prerequisites

* Install kubectl and have an appropriate kubeconfig entry to your Kubernetes cluster.
* Install kpt.  
    `$ gcloud components install kpt`
* Download the helm chart to your filesystem.  
    `$ git clone -q https://https://github.com/bitnami/charts.git`

#### Steps

1. Run the helm kpt function on each of the charts you need. You can pipe these executions, as shown below. The following commands expand the mongodb and redis charts into a new output directory.  
`$ mkdir output`  
`$ docker run -v $(pwd)/charts/bitnami:/source gcr.io/kpt-functions/helm-template chart_path=/source/mongodb name=my-mongodb | docker run -i -v $(pwd)/charts/bitnami:/source gcr.io/kpt-functions/helm-template name=my-redis chart_path=/source/redis | kpt fn sink output`

2. See a summary of this operation using `kpt config tree`.  
`$ kpt fn source output | kpt config tree`

3. Apply these configs to a kubernetes cluster.  
`$ kubectl apply -R -f ./output`

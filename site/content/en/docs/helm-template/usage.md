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

# Using helm-template

Kpt packages are just configuration so any solution, like the `helm template` command, which emits configuration can also be used to generate kpt packages. The `helm-template` kpt function generates a new kpt package from a local Helm chart or upserts Helm chart configuration to an existing kpt package. In the context of a pipeline, these packages can then be further customized using other kpt functions.

## Examples

### Example 1: Hello World

#### Prerequisites

* Install kubectl and have an appropriate kubeconfig entry to your Kubernetes cluster.
* Install kpt.  

    ```sh
    gcloud components install kpt
    ```

* Install helm.

#### Steps

1. Create a new helm chart called "helloworld-chart".  

    ```sh
    helm create helloworld-chart
    ```

1. Run `helm-template` to expand "helloworld-chart" using name "my-first-example" and see the configuration in a ResourceList.  

    ```sh
    docker run --mount type=bind,source=$(pwd),destination=/source gcr.io/kpt-functions/helm-template chart_path=/source/helloworld-chart name=my-first-example
    ```

1. Save the expanded configuration locally as yaml files by piping through `kpt fn sink`.  

    ```sh
    mkdir helloworld-configs
    docker run --mount type=bind,source=$(pwd),destination=/source gcr.io/kpt-functions/helm-template chart_path=/source/helloworld-chart name=my-first-example |
    kpt fn sink helloworld-configs
    ```

### Example 2: Expand and apply multiple charts to a cluster

#### Prerequisites

* Install kubectl and have an appropriate kubeconfig entry to your Kubernetes cluster.
* Install kpt.  

    ```sh
    gcloud components install kpt
    ```

* Install helm.
* Download the helm charts for this example to your filesystem or use your own.  

    ```sh
    helm repo add bitnami https://charts.bitnami.com/bitnami
    helm pull bitnami/mongodb --untar
    helm pull bitnami/redis --untar
    ```

#### Steps

1. Run `helm-template` on each of the charts you need. You can pipe these commands, as shown below. The following commands expand the mongodb and redis charts and store the resulting yaml into a new output directory.  

    ```sh
    mkdir output
    docker run --mount type=bind,source=$(pwd),destination=/source gcr.io/kpt-functions/helm-template chart_path=/source/mongodb name=my-mongodb |
    docker run -i --mount type=bind,source=$(pwd),destination=/source gcr.io/kpt-functions/helm-template name=my-redis chart_path=/source/redis |
    kpt fn sink output
    ```

2. See a summary of the output using `kpt config tree`.  

    ```sh
    kpt fn source output |
    kpt config tree
    ```

3. Apply these configs to a kubernetes cluster.  

    ```sh
    kubectl apply -R -f output
    ```

## FAQs

### How can I set arbitrary values in my chart using `--set`

We recommend that you create a new values.yaml file with the values you want so you can check the new file into a version-controlled repository. You can specify an optional `values_path` argument to the helm-template command containing the relative path to your new file.  

```sh
docker run --mount type=bind,source=$(pwd),destination=/source gcr.io/kpt-functions/helm-template values_path=/source/redis/values-production.yaml chart_path=/source/redis name=my-redis
```

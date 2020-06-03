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
    docker run -u "$(id -u)" --mount type=bind,source=$(pwd),destination=/source gcr.io/kpt-functions/helm-template -d name=my-first-example -d chart_path=/source/helloworld-chart
    ```

1. Save the expanded configuration locally as yaml files by piping through `kpt fn sink`.  

    ```sh
    mkdir helloworld-configs
    docker run -u "$(id -u)" --mount type=bind,source=$(pwd),destination=/source gcr.io/kpt-functions/helm-template -d name=my-first-example -d chart_path=/source/helloworld-chart |
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
    docker run -u "$(id -u)" --mount type=bind,source=$(pwd),destination=/source gcr.io/kpt-functions/helm-template -d name=my-mongodb -d chart_path=/source/mongodb |
    docker run -i -u "$(id -u)" --mount type=bind,source=$(pwd),destination=/source gcr.io/kpt-functions/helm-template -d name=my-redis -d chart_path=/source/redis |
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

### How can I set arbitrary values in my chart

We recommend that you create a function config file with the values you want so you can check the new file into a version-controlled repository. You can specify arbitrary arguments to the helm-template command in this way. The below example specifies a different values file than the default. 

```sh
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
data:
  name: my-prod-redis
  chart_path: /source/redis
  --values: /source/redis/values-production.yaml
metadata:
  name: my-function-config
EOF
docker run -u "$(id -u)" --mount type=bind,source=$(pwd),destination=/source gcr.io/kpt-functions/helm-template -f /source/fc.yaml
```

# inflate-helm-chart: Local Chart

### Overview

This example demonstrated how to imperatively invoke the `inflate-helm-chart`
function to inflate a helm chart that lives in your local filesystem.

### Function invocation

Run the following command to fetch the example package:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/inflate-helm-chart-local
```

```shell
$ cd inflate-helm-chart-local
```

Run the following commands to inflate the helm chart in your local
filesystem.

```shell
$ kpt fn eval --image gcr.io/kpt-fn/inflate-helm-chart:unstable \
--mount type=bind,src=$(pwd),dst=/tmp/charts \
-- name=helloworld-chart \
releaseName=test
```

You can optionally provide your own values files using `--valuesFile`.

```shell
$ kpt fn eval --image gcr.io/kpt-fn/inflate-helm-chart:unstable \
--mount type=bind,src=$(pwd),dst=/tmp/charts -- \
name=helloworld-chart \
releaseName=test \
valuesFile=tmp/charts/helloworld-values/values.yaml
```

### Expected result

You can run the following command to see the new files you have:

```shell
$ kpt pkg tree
├── [deployment_test-helloworld-chart.yaml]  Deployment test-helloworld-chart
├── [pod_test-helloworld-chart-test-connection.yaml]  Pod test-helloworld-chart-test-connection
├── [service_test-helloworld-chart.yaml]  Service test-helloworld-chart
└── [serviceaccount_test-helloworld-chart.yaml]  ServiceAccount test-helloworld-chart
```

You should be able to find `replicas: 5` in
file `deployment_test-helloworld-chart.yaml`, which demonstrates that
the correct values file provided by --valuesFile was used.

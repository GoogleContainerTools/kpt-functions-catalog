# inflate-helm-chart: Local Chart

### Overview

This example demonstrated how to imperatively invoke the `inflate-helm-chart`
function to inflate a helm chart that lives in your local filesystem.

### Function invocation

Run the following command to fetch the example package:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/contrib/inflate-helm-chart/local-chart
```

```shell
$ cd local-chart
```

Run the following commands to inflate a nginx helm chart from your local
filesystem.

```shell
$ kpt fn eval --image gcr.io/kpt-fn-contrib/inflate-helm-chart:unstable \
--mount type=bind,src=$(pwd)/helloworld-chart,dst=/source -- \
name=helloworld \
local-chart-path=/source \
```

You can optionally provide your own values files using `--value`.

```shell
$ kpt fn eval --image gcr.io/kpt-fn-contrib/inflate-helm-chart:unstable \
--mount type=bind,src=$(pwd)/helloworld-chart,dst=/source \
--mount type=bind,src=$(pwd)/helloworld-values,dst=/values -- \
name=helloworld \
local-chart-path=/source \
--values=/values/values.yaml
```

### Expected result

You should have several files in your local filesystem. You can run the
following command to see what you have:

```shell
$ kpt pkg tree
├── [deployment_nginx-helloworld-chart.yaml]  Deployment nginx-helloworld-chart
├── [pod_nginx-helloworld-chart-test-connection.yaml]  Pod nginx-helloworld-chart-test-connection
├── [service_nginx-helloworld-chart.yaml]  Service nginx-helloworld-chart
└── [serviceaccount_nginx-helloworld-chart.yaml]  ServiceAccount nginx-helloworld-chart
```

You should be able to find `replicas: 5` in
file `deployment_nginx-helloworld-chart.yaml`, which mean the values file
provided using `--values` took effects.

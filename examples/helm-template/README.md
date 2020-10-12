# Helm Template

The `helm-template` KRM config function generates a new kpt package from a
local Helm chart. This example invokes the helm template function using
declarative configuration.

## Function invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/helm-template .
kpt fn run helm-template/local-configs --mount type=bind,src=$(pwd)/helm-template/helloworld-chart,dst=/source
```

## Expected result

Checking the contents of the `local-configs` directory with `kpt cfg tree helm-template/local-configs/` should reveal the following new yaml files:

```sh
helm-template/local-configs
├── [deployment_chart-helloworld-chart.yaml]  Deployment chart-helloworld-chart
├── [fn-config.yaml]  ConfigMap my-func-config
├── [pod_chart-helloworld-chart-test-connection.yaml]  Pod chart-helloworld-chart-test-connection
├── [service_chart-helloworld-chart.yaml]  Service chart-helloworld-chart
└── [serviceaccount_chart-helloworld-chart.yaml]  ServiceAccount chart-helloworld-chart
```

To view changes without writing them into a file, a dry run can be performed as follows:

```sh
kpt fn run helm-template/local-configs --mount type=bind,src=$(pwd)/helm-template/helloworld-chart,dst=/source --dry-run
```

The expected output should match the following:

```yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: chart-helloworld-chart
  labels:
    helm.sh/chart: helloworld-chart-0.1.0
    app.kubernetes.io/name: helloworld-chart
    app.kubernetes.io/instance: chart
    app.kubernetes.io/version: 1.16.0
    app.kubernetes.io/managed-by: Helm
  annotations:
    config.kubernetes.io/path: 'deployment_chart-helloworld-chart.yaml'
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: helloworld-chart
      app.kubernetes.io/instance: chart
  template:
    metadata:
      labels:
        app.kubernetes.io/name: helloworld-chart
        app.kubernetes.io/instance: chart
    spec:
      serviceAccountName: chart-helloworld-chart
      securityContext: {}
      containers:
      - name: helloworld-chart
        securityContext: {}
        image: 'nginx:1.16.0'
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          containerPort: 80
          protocol: TCP
        livenessProbe:
          httpGet:
            path: /
            port: http
        readinessProbe:
          httpGet:
            path: /
            port: http
        resources: {}
---
# call `kpt fn run` on a directory containing this file, mounting the helm chart at /source
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-func-config
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/kpt-functions/helm-template
    config.kubernetes.io/path: fn-config.yaml
data:
  name: chart
  chart_path: /source
---
apiVersion: v1
kind: Pod
metadata:
  name: chart-helloworld-chart-test-connection
  labels:
    helm.sh/chart: helloworld-chart-0.1.0
    app.kubernetes.io/name: helloworld-chart
    app.kubernetes.io/instance: chart
    app.kubernetes.io/version: 1.16.0
    app.kubernetes.io/managed-by: Helm
  annotations:
    helm.sh/hook: test-success
    config.kubernetes.io/path: 'pod_chart-helloworld-chart-test-connection.yaml'
spec:
  containers:
  - name: wget
    image: busybox
    command:
    - wget
    args:
    - 'chart-helloworld-chart:80'
  restartPolicy: Never
---
apiVersion: v1
kind: Service
metadata:
  name: chart-helloworld-chart
  labels:
    helm.sh/chart: helloworld-chart-0.1.0
    app.kubernetes.io/name: helloworld-chart
    app.kubernetes.io/instance: chart
    app.kubernetes.io/version: 1.16.0
    app.kubernetes.io/managed-by: Helm
  annotations:
    config.kubernetes.io/path: 'service_chart-helloworld-chart.yaml'
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: http
    protocol: TCP
    name: http
  selector:
    app.kubernetes.io/name: helloworld-chart
    app.kubernetes.io/instance: chart
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: chart-helloworld-chart
  labels:
    helm.sh/chart: helloworld-chart-0.1.0
    app.kubernetes.io/name: helloworld-chart
    app.kubernetes.io/instance: chart
    app.kubernetes.io/version: 1.16.0
    app.kubernetes.io/managed-by: Helm
  annotations:
    config.kubernetes.io/path: 'serviceaccount_chart-helloworld-chart.yaml'
```

# inflate-helm-chart: Remote Chart

### Overview

This example demonstrated how to imperatively invoke the `inflate-helm-chart`
function to inflate a helm chart that lives in a remote repo.

### Function invocation

Run the following command to inflate a nginx helm chart from bitnami's repo.

```shell
$ kpt fn eval --image gcr.io/kpt-fn-contrib/inflate-helm-chart:unstable --network -- \
name=nginx \
chart=bitnami/nginx \
chart-repo=bitnami \
chart-repo-url=https://charts.bitnami.com/bitnami \
chart-version=8.4.0
```

### Expected result

You should have several files in your local filesystem. You can run the
following command to see what you have:

```shell
$ kpt pkg tree
├── [configmap_nginx-server-block.yaml]  ConfigMap nginx-server-block
├── [deployment_nginx.yaml]  Deployment nginx
└── [service_nginx.yaml]  Service nginx
```

# render-helm-chart: Kustomize Values Files

### Overview

This example demonstrates how to declaratively invoke the `render-helm-chart`
function with kustomize using multiple values files.

### Function invocation

To use the function with kustomize, you can specify the `functionConfig`
in your kustomization's `generators` field. This example specifies multiple remote
values files to use instead of the default values accompanying the chart:

kustomization.yaml:
```yaml
generators:
- |-
  apiVersion: fn.kpt.dev/v1alpha1
  kind: RenderHelmChart
  metadata:
    name: demo
    annotations:
      config.kubernetes.io/function: |
        container:
          network: true
          image: gcr.io/kpt-fn/render-helm-chart:unstable
  helmCharts:
  - chartArgs:
      name: ocp-pipeline
      version: 0.1.16
      repo: https://bcgov.github.io/helm-charts
    templateOptions:
      namespace: mynamespace
      releaseName: moria
      values:
        valuesFiles:
          - https://raw.githubusercontent.com/natasha41575/kpt-functions-catalog/a9c9cd765a05f7a7fb6923dbde4651b62c9c229c/examples/render-helm-chart-kustomize-values-files/file1.yaml
          - https://raw.githubusercontent.com/natasha41575/kpt-functions-catalog/a9c9cd765a05f7a7fb6923dbde4651b62c9c229c/examples/render-helm-chart-kustomize-values-files/file2.yaml
```

Then, to build the kustomization with kustomize v4:

```shell
kustomize build --enable-alpha-plugins --network .
```

### Expected result

You should also be able to find the line `def releaseNamespace = ""` somewhere
in your output, as well as the following: 

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: moria-ocp-pipeline
  namespace: mynamespace
rules:
- apiGroups:
  - ""
  resources:
  - '*'
  verbs:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: moria-ocp-pipeline
  namespace: mynamespace
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: moria-ocp-pipeline
subjects:
- kind: ServiceAccount
  name: jenkins
  namespace: mynamespace
```

which demonstrates that the correct values provided via `valuesFiles` were used.

# render-helm-chart: Kustomize Private OCI Registry Chart

### Overview

This example demonstrates how to declaratively invoke the `render-helm-chart`
function with kustomize with charts hosted in a private OCI registry.

### Function invocation

To use the function with kustomize, you can specify the `functionConfig`
in your kustomization's `generators` field and create the Secret holding
the private repo credentials using the `secretGenerator`:

kustomization.yaml:
```yaml
secretGenerator:
  - name: my-secret
    literals:
      - username=_json_key
    files:
      - password
    options:
      disableNameSuffixHash: true
      annotations:
        config.kubernetes.io/local-config: "true"

transformers:
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
        name: mychart # change this to the name of your chart
        repo: oci://us-west2-docker.pkg.dev/kustomize-326923/helloworld-chart # change this to your private OCI repo
        registry: https://us-west2-docker.pkg.dev # change this to your OCI registry
        auth:
          kind: Secret
          name: my-secret
```

Then, to build the kustomization with kustomize v4:

```shell
kustomize build --enable-alpha-plugins --network .
```

### Expected result

You should see your inflated private chart in the output, demonstrating that the function was able to
use the provided credentials to access your chart stored in a private OCI registry.

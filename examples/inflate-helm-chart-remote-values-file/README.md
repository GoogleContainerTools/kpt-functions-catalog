# inflate-helm-chart: Remote Values File

### Overview

This example demonstrates how to imperatively invoke the `inflate-helm-chart`
function with a remote values file.

### Function invocation

Run the following command to fetch the example package:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/inflate-helm-chart-remote-values-file
```

```shell
$ cd inflate-helm-chart-remote-values-file
```

You can inflate the helm chart using a remote values file with the following command:

```shell
$ kpt fn eval --image gcr.io/kpt-fn/inflate-helm-chart:unstable \
  --network \
  --mount type=bind,src="$(pwd)",dst=/tmp/charts -- \
  name=cert-manager \
  namespace=cert-manager \
  releaseName=cert-manager \
  valuesFile=https://raw.githubusercontent.com/config-sync-examples/helm-components/main/cert-manager-values.yaml
```

### Expected result

You should have a number of new files in your local file system. You can view the contents of your current package:

```shell
$ kpt pkg tree
├── [clusterrole_cert-manager-cainjector.yaml]  ClusterRole cert-manager-cainjector
├── [clusterrole_cert-manager-controller-approve:cert-manager-io.yaml]  ClusterRole cert-manager-controller-approve:cert-manager-io
├── [clusterrole_cert-manager-controller-certificates.yaml]  ClusterRole cert-manager-controller-certificates
├── [clusterrole_cert-manager-controller-challenges.yaml]  ClusterRole cert-manager-controller-challenges
├── [clusterrole_cert-manager-controller-clusterissuers.yaml]  ClusterRole cert-manager-controller-clusterissuers
├── [clusterrole_cert-manager-controller-ingress-shim.yaml]  ClusterRole cert-manager-controller-ingress-shim
├── [clusterrole_cert-manager-controller-issuers.yaml]  ClusterRole cert-manager-controller-issuers
├── [clusterrole_cert-manager-controller-orders.yaml]  ClusterRole cert-manager-controller-orders
├── [clusterrole_cert-manager-edit.yaml]  ClusterRole cert-manager-edit
├── [clusterrole_cert-manager-view.yaml]  ClusterRole cert-manager-view
├── [clusterrole_cert-manager-webhook:subjectaccessreviews.yaml]  ClusterRole cert-manager-webhook:subjectaccessreviews
├── [clusterrolebinding_cert-manager-cainjector.yaml]  ClusterRoleBinding cert-manager-cainjector
├── [clusterrolebinding_cert-manager-controller-approve:cert-manager-io.yaml]  ClusterRoleBinding cert-manager-controller-approve:cert-manager-io
├── [clusterrolebinding_cert-manager-controller-certificates.yaml]  ClusterRoleBinding cert-manager-controller-certificates
├── [clusterrolebinding_cert-manager-controller-challenges.yaml]  ClusterRoleBinding cert-manager-controller-challenges
├── [clusterrolebinding_cert-manager-controller-clusterissuers.yaml]  ClusterRoleBinding cert-manager-controller-clusterissuers
├── [clusterrolebinding_cert-manager-controller-ingress-shim.yaml]  ClusterRoleBinding cert-manager-controller-ingress-shim
├── [clusterrolebinding_cert-manager-controller-issuers.yaml]  ClusterRoleBinding cert-manager-controller-issuers
├── [clusterrolebinding_cert-manager-controller-orders.yaml]  ClusterRoleBinding cert-manager-controller-orders
├── [clusterrolebinding_cert-manager-webhook:subjectaccessreviews.yaml]  ClusterRoleBinding cert-manager-webhook:subjectaccessreviews
├── [mutatingwebhookconfiguration_cert-manager-webhook.yaml]  MutatingWebhookConfiguration cert-manager-webhook
├── [validatingwebhookconfiguration_cert-manager-webhook.yaml]  ValidatingWebhookConfiguration cert-manager-webhook
└── kube-system
    ├── [role_cert-manager-cainjector:leaderelection.yaml]  Role kube-system/cert-manager-cainjector:leaderelection
    ├── [role_cert-manager:leaderelection.yaml]  Role kube-system/cert-manager:leaderelection
    ├── [rolebinding_cert-manager-cainjector:leaderelection.yaml]  RoleBinding kube-system/cert-manager-cainjector:leaderelection
    └── [rolebinding_cert-manager:leaderelection.yaml]  RoleBinding kube-system/cert-manager:leaderelection
```

In `cert-manager/deployment_cert-manager-cainjector.yaml` and `cert-manager/deployment_cert-manager-cainjector.yaml`
you should be able to find `imagePullPolicy: Always`, which indicates that the remote values file was used.
If you run the command without the remote values file, you will find `imagePullPolicy: IfNotPresent` instead.
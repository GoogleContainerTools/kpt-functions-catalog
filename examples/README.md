# KPT Function Catalog

## Mutators

| Name            | Description |
| --------------- | ----------- |
| [Helm Inflator](/helm-inflator/latest/)| Render chart templates locally using helm template. |
| Set Annotation | N/A |
| Set Label | N/A |
| Set Namespace | Sets the namespace field of all configs passed in. |
| Sops | N/A |

## Validators

| Name            | Description |
| --------------- | ----------- |
| Istioctl Analyze | A diagnostic tool that can detect potential issues with Istio configuration and output errors to the results field. |
| Kubeval | Validates configuration using kubeval. |
| Suggest PSP | [Demo] Lints PodSecurityPolicy by suggesting ‘spec.allowPrivilegeEscalation’ field be set to ‘false’. |
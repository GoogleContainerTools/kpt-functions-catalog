# OpenAPI

`k8s.json` should be the OpenAPI schema downloaded from a GKE cluster with
Config Connector resources installed. To download an OpenAPI schema file:

```shell
curl -k -H "Authorization: Bearer $(gcloud auth print-access-token)"  https://${your-apiserver-address}/openapi/v2 | jq > k8s.json
```

# Merge kptfile openapi schema

Download the openapi schema from the kpt-functions-sdk repo and then convert it
to json format.

```shell
curl https://raw.githubusercontent.com/GoogleContainerTools/kpt-functions-sdk/master/openapi/kptfile.yaml | yq eval -o json > kptfile.json
```

Merge the kptfile openapi schema with the other k8s openapi schema

```shell
jq -s '.[0] * .[1]' k8s.json kptfile.json > openapi.json
```

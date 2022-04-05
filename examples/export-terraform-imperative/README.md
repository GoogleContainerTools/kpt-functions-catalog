# export-terraform: Imperative Example

### Overview

In this example, we will see how to export Terraform configuration from KCC resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/export-terraform-imperative@export-terraform/v0.1.0
```

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn eval export-terraform-imperative --image gcr.io/kpt-fn/export-terraform:v0.1.0
```

### Expected result
The function should export successfully
```shell
[RUNNING] "gcr.io/kpt-fn/export-terraform:v0.1.0"
[PASS] "gcr.io/kpt-fn/export-terraform:v0.1.0" in 1.5s
```

A `ConfigMap` will be placed in `terraform.yaml` which contains the converted Terraform code.

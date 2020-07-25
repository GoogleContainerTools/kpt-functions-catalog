# Set Namespace Config Function

This function sets the metadata.namespace field on all resources.

## One time setup

```sh
go mod init
go get sigs.k8s.io/kustomize/kyaml
```

## Build the binary

```sh
go build .
```

## Test the binary

```sh
kpt fn source config/ | kpt fn run --enable-exec --exec-path ./set-namespace -- namespace=test-ns
```

## Generate a Dockerfile to contain the function

```sh
go run ./main.go gen ./
```

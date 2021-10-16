# preview-hierarchy: Simple Example

### Overview

In this example, we will see how to print Folder hierarchy to stdout for a 
provided set of Folder resources

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/preview-hierarchy-simple@preview-hierarchy/v0.1
```

resources.yaml in this package will be the input resource list for this function invocation

### Function invocation

Invoke the function by running the following command:

```shell
$ kpt fn eval preview-hierarchy-simple -i gcr.io/kpt-fn/preview-hierarchy:v0.1 --results-dir /tmp/preview-hierarchy -- renderer=text
```

### Expected result

The generated file: `/tmp/preview-hierarchy/results.yaml` should have the following output:

```shell
apiVersion: kpt.dev/v1
kind: FunctionResultList
metadata:
  name: fnresults
exitCode: 0
items:
  - image: gcr.io/kpt-fn/preview-hierarchy:v0.1
    stderr: |2

      org-11111
      ├─Commercial
      ├─Financial
      | ├─Apps
      | | ├─Dev
      | | ├─Prod
      | | └─Test
      | ├─Shared
      | └─Web
      ├─Retail
      | ├─Apps
      | | ├─Dev
      | | ├─Prod
      | | └─Test
      | ├─Shared
      | └─Web
      └─Risk Mgmt
    exitCode: 0
```

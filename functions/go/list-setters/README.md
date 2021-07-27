# list-setters

## Overview

<!--mdtogo:Short-->

Lists information about [setters] like setter name, values and count.

Refer to the create-setters function documentation for information about creating new setters or apply-setters function documentation for information about parameterizing field values using setters.

<!--mdtogo-->

<!--mdtogo:Long-->

## Usage

`list-setters` function is expected to be executed imperatively like

```shell
$ kpt fn eval -i list-setters:unstable --include-meta-resources
```

`list-setters` function performs the following steps:

1. Searches for apply-setters functionConfig in the Kptfile (if present) for setter information.
1. Searches for setter comments in input list of resources.
1. Lists discovered setters and related information.

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

### Listing setters with Kptfile

Let's start with an input resource in a package with a Kptfile

```yaml
# Kptfile
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: nginx
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configPath: setters.yaml
```

```yaml
# setters.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: apply-setters-fn-config
data:
  env: |
      - dev
      - stage
  unused: notused
```

```yaml
# resources.yaml
apiVersion: v1
kind: MyKind
metadata:
  name: foo
environments: # kpt-set: ${env}
  - dev
  - stage
```

Invoke the function:

```shell
$ kpt fn eval --image gcr.io/kpt-fn/list-setters:unstable --include-meta-resources
```

Output looks like the following:

```shell
  Results:
    [INFO] Name: env, Value: [dev, stage], Type: list, Count: 1
    [INFO] Name: unused, Value: notused, Type: string, Count: 0
```


### Listing setters in simple config

Let's start with the input resource in a package without a Kptfile

```yaml
# resources.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-nginx
spec:
  replicas: 4 # kpt-set: ${nginx-replicas}
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: "nginx:1.16.1" # kpt-set: nginx:${tag}
        ports:
        - protocol: TCP
          containerPort: 80
```

Invoke the function:

```shell
$ kpt fn eval --image gcr.io/kpt-fn/list-setters:unstable
```

Output looks like the following:

```shell
  Results:
    [WARNING] unable to find Kptfile, please include --include-meta-resources flag if a Kptfile is present
    [INFO] Name: nginx-replicas, Value: 4, Type: string, Count: 1
    [INFO] Name: tag, Value: 1.16.1, Type: string, Count: 1
```

<!--mdtogo-->

[setter]: https://catalog.kpt.dev/apply-setters/v0.1/?id=definitions
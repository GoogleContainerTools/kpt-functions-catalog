# format

### Overview

<!--mdtogo:Short-->

Format resources in a directory.

<!--mdtogo-->

### Synopsis

<!--mdtogo:Long-->

The `format` function formats the field ordering in YAML configuration files. Field
ordering follows the ordering defined in the source Kubernetes resource definitions,
falling back on lexicographical sorting for unrecognized fields.

<!--mdtogo-->

### Examples

<!--mdtogo:Examples-->

#### Format a package

Let's start with the input resource in a package.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    bar: 100
  name: nginx-deployment
spec:
  template:
    spec:
      containers:
        - name: nginx
          image: nginx:1.0.0
          ports:
            - containerPort: 80
              name: http
```

Invoke the `format` function on the package.

```sh
$ kpt fn run . --image gcr.io/kpt-fn/format:v0.1
```

Formatted resource looks like the following:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  annotations:
    bar: "100"
spec:
  template:
    spec:
      containers:
        - name: nginx
          image: nginx:1.0.0
          ports:
            - name: http
              containerPort: 80
```

<!--mdtogo-->

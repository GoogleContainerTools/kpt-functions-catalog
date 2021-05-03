# format

### Overview

<!--mdtogo:Short-->

Format the field ordering in resources.

<!--mdtogo-->

### Synopsis

<!--mdtogo:Long-->

The `format` function formats the field ordering in YAML configuration files. Field
ordering follows the ordering defined in the openapi document for Kubernetes resources,
falling back on lexicographical sorting for unrecognized fields. This function also performs
other changes like fixing indentation, adding quotes to ambiguous string values.

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

The fields are ordered as per the openapi schema definition of `Deployment` resource. For e.g. `metadata.name` field
is moved up. Since the type of annotation value is `string`, quotes are added to value `100` as it will be interpreted
as `int` by yaml in its current form.

<!--mdtogo-->

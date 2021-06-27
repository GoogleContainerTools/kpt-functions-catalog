# format

### Overview

<!--mdtogo:Short-->

Sort resource fields, fix indentation and quote ambiguous string values.

<!--mdtogo-->

### FunctionConfig

<!--mdtogo:Long-->

The `format` function sorts the resource fields. Field ordering follows the 
ordering defined in the OpenAPI document for Kubernetes resources,
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

Invoke the `format` function on the package, formatted resource looks like the following:

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

The fields are ordered as per the OpenAPI schema definition of `Deployment` resource. For e.g. `metadata.name` field
is moved up. Since the type of annotation value is `string`, quotes are added to value `100` as it will be interpreted
as `int` by yaml in its current form.

<!--mdtogo-->

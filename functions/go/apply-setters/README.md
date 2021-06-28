# apply-setters

### Overview

<!--mdtogo:Short-->

Apply setter values on resource fields. Setters serve as parameters for template-free
setting of field values.

Setters are a safer alternative to parameterize field values as they
have the context of the structured data. Setters can be invoked to modify the
resources using this function to set desired values.

<!--mdtogo-->

### FunctionConfig

<!--mdtogo:Long-->

We use ConfigMap to configure the `apply-setters` function. The desired setter
values are provided as key-value pairs using `data` field where key is the name of the
setter and value is the new desired value for the setter.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: apply-setters-func-config
data:
  setter_name1: setter_value1
  setter_name2: setter_value2
```

<!--mdtogo-->

### Examples

<!--mdtogo:Examples-->

#### Setting scalar values

Let's start with the input resource in a package

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

Declare the new desired values for setters in the function config file.

```yaml
# apply-setters-fn-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: apply-setters-fn-config
data:
  nginx-replicas: "3"
  tag: 1.16.2
```

Invoke the function using the input config:

```shell
$ kpt fn eval --image gcr.io/kpt-fn/apply-setters:unstable --fn-config ./apply-setters-fn-config
```

Alternatively, setter values can be passed as key-value pairs in the CLI

```shell
$ kpt fn eval --image gcr.io/kpt-fn/apply-setters:unstable -- image=ubuntu replicas=3
```

Modified resource looks like the following:

```yaml
# resources.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-nginx
spec:
  replicas: 3 # kpt-set: ${nginx-replicas}
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
        image: "nginx:1.16.2" # kpt-set: nginx:${tag}
        ports:
        - protocol: TCP
          containerPort: 80
```

#### Setting array values

Array values can also be parameterized using setters. Since the values of configMap
in pipeline definition must be of string type, the array values must be wrapped into
string. However, the rendered values in the resources will be array type.

Let's start with the input resource

```yaml
apiVersion: v1
kind: MyKind
metadata:
  name: foo
environments: # kpt-set: ${env}
  - dev
  - stage
```

Declare the desired array values, wrapped into string.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: apply-setters-fn-config
data:
  env: |
    - prod
    - dev
```

Invoke the function using the input config:

```shell
$ kpt fn eval --image gcr.io/kpt-fn/apply-setters:unstable --fn-config ./apply-setters-fn-config
```

Modified resource looks like the following:

```yaml
apiVersion: v1
kind: MyKind
metadata:
  name: foo
environments: # kpt-set: ${env}
  - prod
  - dev
```

<!--mdtogo-->

#### Note:

Refer to the `create-setters` function documentation for information about creating setters.

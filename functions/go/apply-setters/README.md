# apply-setters

### Overview

<!--mdtogo:Short-->

Apply setter values on resource fields. Setters serve as parameters for template-free
setting of field values.

Setters are a safer alternative to other substitution techniques which do not
have the context of the structured data. Setters may be invoked to modify the
package resources using this function to set desired values.

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
  name: my-func-config
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
apiVersion: v1
kind: Deployment
metadata:
  name: nginx-deployment # kpt-set: ${image}-deployment
spec:
  replicas: 1 # kpt-set: ${replicas}
```

Discover the names of setters in the function config file and declare desired values.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: apply-setters-fn-config
data:
  image: ubuntu
  replicas: "3"
```

Render the declared values by invoking:

```shell
$ kpt fn eval --image gcr.io/kpt-fn/apply-setters:unstable --fn-config ./apply-setters-fn-config
```

Alternatively, setter values can be passed as key-value pairs in the CLI

```shell
$ kpt fn eval --image gcr.io/kpt-fn/apply-setters:unstable -- 'image=ubuntu' 'replicas=3'
```

Rendered resource looks like the following:

```yaml
apiVersion: v1
kind: Deployment
metadata:
  name: ubuntu-deployment # kpt-set: ${image}-deployment
spec:
  replicas: 3 # kpt-set: ${replicas}
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

Render the declared values by invoking:

```
$ kpt fn eval --image gcr.io/kpt-fn/apply-setters:unstable --fn-config ./apply-setters-fn-config
```

Rendered resource looks like the following:

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

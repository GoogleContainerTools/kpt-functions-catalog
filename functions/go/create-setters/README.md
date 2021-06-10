# create-setters

### Overview

<!--mdtogo:Short-->

Add setter comments to matching resource fields. Setters serve as parameters 
for template-free setting of comments.

Setters are a safer alternative to other substitution techniques which do not
have the context of the structured data. Setter comments can be added to
parameterize the field values of resources using this function.

>? Refer to [`apply-setters`](https://catalog.kpt.dev/apply-setters/v0.1/) for easy understanding of setters.

<!--mdtogo-->

### FunctionConfig

<!--mdtogo:Long-->

We use `ConfigMap` to configure the `create-setters` function. The desired setter
values are provided as key-value pairs using `data` field.
Here, the key is the name of the setter used as parameter and
value is the field value to parameterize.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: create-setters-fn-config
data:
  setter_name1: setter_value1
  setter_name2: setter_value2
```

### How comments are added
1. Searches for the setter values to be parameterized in each of the resource fields
2. Adds comments to the fields matching the setter values using setter names as parameters

### Adds comment
- Scalar value matching atleast one of the setter values.
- Array values matching all of the values in a setter_value.
- Any of the array value matching any of the setter_value.

### Doesn't add comment
- Resource field values with multiple lines.

>? If this function adds setter comments to fields for which you didn't intend 
to parameterize, you can simply review and delete/modify those comments manually.

<!--mdtogo-->

### Examples

<!--mdtogo:Examples-->

### Setting comment for scalar values

Let's start with the input resource in a package

```yaml
# resources.yaml
apiVersion: v1
kind: Deployment
metadata:
  name: ubuntu-deployment 
spec:
  replicas: 3 
```

Declare the name of the setter with the value which need to be parameterized.

```yaml
# create-setters-fn-config
apiVersion: v1
kind: ConfigMap
metadata:
  name: create-setters-fn-config
data:
  image: ubuntu
  replicas: "3"
```

Render the declared values by invoking:

```shell
$ kpt fn eval --image gcr.io/kpt-fn/create-setters:unstable --fn-config ./create-setters-fn-config
```

Alternatively, setter values can be passed as key-value pairs in the CLI

```shell
$ kpt fn eval --image gcr.io/kpt-fn/create-setters:unstable -- image=ubuntu replicas=3
```

Rendered resource looks like the following:

```yaml
# resources.yaml
apiVersion: v1
kind: Deployment
metadata:
  name: ubuntu-deployment # kpt-set: ${image}-deployment
spec:
  replicas: 3 # kpt-set: ${replicas}
```
- As `metadata.name` field value contains a match with the setter value of `image`, `# kpt-set: ${image}-development` comment is added
- As `spec.replicas` field value matches with the setter value of `replicas`, `# kpt-set: ${replicas}` comment is added


### Setting comment for array values

Fields with array values can also be parameterized using setters. Since the values of `ConfigMap`
in pipeline definition must be of string type, the array values must be wrapped into
string.

Let's start with the input resource

```yaml
# resources.yaml
apiVersion: v1
kind: MyKind
metadata:
  name: foo
environments:
  - dev
  - stage
```

Declare the array values, wrapped into string. Here the order of the array values
doesn't make a difference.

```yaml
# create-setters-fn-config
apiVersion: v1
kind: ConfigMap
metadata:
  name: create-setters-fn-config
data:
  role: dev
  env: |
    - dev
    - stage
```

Render the declared values by invoking:

```shell
$ kpt fn eval --image gcr.io/kpt-fn/create-setters:unstable --fn-config ./create-setters-fn-config
```

Rendered resource looks like the following:

```yaml
# resources.yaml
apiVersion: v1
kind: MyKind
metadata:
  name: foo
environments: # kpt-set: ${env}
  - dev # kpt-set: ${role}
  - stage
```

- As the values for `environments` match the setter values of `env`, `# kpt-set: ${env}` comment is added.
- As the array value `dev` matches with the setter value of `role`, `# kpt-set: ${role}` comment is added.
<!--mdtogo-->

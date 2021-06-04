# create-setters

### Overview

<!--mdtogo:Short-->

Add setter comments to matching resource fields. Setters serve as parameters 
for template-free setting of comments.

<!--mdtogo-->

### FunctionConfig

<!--mdtogo:Long-->

Setters are a safer alternative to other substitution techniques which do not
have the context of the structured data. Setter comments can be added to
parameterize the field values of resources using create-setters function.

We use ConfigMap to configure the `create-setters` function. The desired setter
values are provided as key-value pairs using `data` field.
Here, the key is the name of the setter which is used to set the comment and
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

<!--mdtogo-->

### Examples

<!--mdtogo:Examples-->

#### Setting comment for scalar values

Let's start with the input resource in a package

```yaml
apiVersion: v1
kind: Deployment
metadata:
  name: ubuntu-deployment 
spec:
  replicas: 3 
```

Declare the name of the setter with the value for which comments should be added.

```yaml
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
apiVersion: v1
kind: Deployment
metadata:
  name: ubuntu-deployment # kpt-set: ${image}-deployment
spec:
  replicas: 3 # kpt-set: ${replicas}
```

#### Setting comment for array values

Array values can also be parameterized using setters. Since the values of configMap
in pipeline definition must be of string type, the array values must be wrapped into
string.

Let's start with the input resource

```yaml
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
apiVersion: v1
kind: ConfigMap
metadata:
  name: create-setters-fn-config
data:
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
apiVersion: v1
kind: MyKind
metadata:
  name: foo
environments: # kpt-set: ${env}
  - dev
  - stage
```
<!--mdtogo-->

#### Note:

If this function adds setter comments to fields for which you didn't intend to parameterize,
you can simply review and delete those comments manually.
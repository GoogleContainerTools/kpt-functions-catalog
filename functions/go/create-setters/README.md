# create-setters

### Overview

<!--mdtogo:Short-->

Add setter comments to matching resource fields. Setters serve as
parameters for template-free setting of field values.

Setters are a safer alternative to other substitution techniques which do not
have the context of the structured data. Setter comments can be added to
parameterize the field values of resources using this function.

>? Refer to `apply-setters` for easy understanding of setters.

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

On invoking `create-setters`, it performs the following steps:
1. Searches for the values to be parameterized in each of the resource fields.
2. Checks if there is a match considering following cases.,
   - Scalar value matching atleast one of the setter values as a substring.
   - Array values matching all of the values in a setter_value.
3. If there are two or more setter values with similar substrings matching the resource field, then longest length matched setter_value is considered.
4. Adds comments to the fields matching the setter values using setter names as parameters.

>? Doesn't support adding comment to the resource field whose value is split into multiple lines.
If this function adds setter comments to fields for which you didn't intend to parameterize,
you can simply review and delete/modify those comments manually.

<!--mdtogo-->

### Examples

<!--mdtogo:Examples-->

### Setting comments for scalar values

Let's start with the input resource in a package

```yaml
# resources.yaml
apiVersion: v1
kind: Deployment
metadata:
  name: ubuntu-deployment 
spec:
  image: ubuntu
  app: nginx:1.1.2
```

Declare the name of the setter with the value which need to be parameterized.

```yaml
# create-setters-fn-config
apiVersion: v1
kind: ConfigMap
metadata:
  name: create-setters-fn-config
data:
  deploy: ubuntu-deployment
  env: ubuntu
  image: ngnix
  tag: 1.1.2
```

Render the declared values by invoking:

```shell
$ kpt fn eval --image gcr.io/kpt-fn/create-setters:unstable --fn-config ./create-setters-fn-config
```

Alternatively, setter values can be passed as key-value pairs in the CLI

```shell
$ kpt fn eval --image gcr.io/kpt-fn/create-setters:unstable -- deploy:ubuntu-deployment env:ubuntu image:ngnix tag:1.1.2
```

Rendered resource looks like the following:

```yaml
# resources.yaml
apiVersion: v1
kind: Deployment
metadata:
  name: ubuntu-deployment-1 # kpt-set: ${deploy}-1
spec:
  image: ubuntu # kpt-set: ${env}
  app: nginx:1.1.2 # kpt-set: ${image}:${tag}
```

Explanation for the changes:
- Value of `metadata.name` matches with setter values of `env` and `deploy` which have same substring `ubuntu`.
As `ubuntu-deployment` has the longest length match, `# kpt-set: ${deploy}-1` comment is added.
- As value of `image` matches with the setter value of `env`, `# kpt-set: ${env}` comment is added.
- As value of `app` matches with setter values of `image` and `tag`, `# kpt-set: ${image}:${tag}` comment is added.

### Setting comments for array values

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

Explanation for the changes:
- As all the values in `environments` matches the setter values of `env`, `# kpt-set: ${env}` comment is added.
- As the array value `dev` matches with the setter value of `role`, `# kpt-set: ${role}` comment is added.

<!--mdtogo-->

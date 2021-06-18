# create-setters

### Overview

<!--mdtogo:Short-->

Add setter comments to matching resource fields. Setters serve as
parameters for the template-free setting of field values.

Setters are a safer alternative to other substitution techniques which do not
have the context of the structured data. Setter comments can be added to
parameterize the field values of resources using this function.

<!--mdtogo-->

### FunctionConfig

<!--mdtogo:Long-->

We use `ConfigMap` to configure the `create-setters` function. The desired setter
values are provided as key-value pairs using `data` field.
Here, the key is the name of the setter used as a parameter and
value is the field value to be parameterized.

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
1. Segregates the input setters into scalar-setters and array-setters.
2. Searches for the resource field values to be parameterized.
3. Checks if there is any match considering the following cases.,
   - For a scalar node, performs substring match with scalar setters.
   - For an array node, checks if all values match with any of the array setters.
4. Adds comments to the fields matching the setter values using setter names as parameters.

>? If this function adds setter comments to fields for which you didn't intend to parameterize,
you can simply review and delete/modify those comments manually.

<!--mdtogo-->

### Examples

<!--mdtogo:Examples-->

### Setting comments for scalar nodes

Let's start with the input resource in a package

```yaml
# resources.yaml
apiVersion: v1
kind: Deployment
metadata:
  name: ubuntu-deployment-1
spec:
  image: ubuntu
  app: "nginx:1.1.2"
  os:
    - ubuntu
    - mac
```

Declare the name of the setter with the value which need to be parameterized.

```yaml
# create-setters-fn-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: create-setters-fn-config
data:
  deploy: ubuntu-deployment
  env: ubuntu
  image: nginx
  tag: 1.1.2
```

Invoke the function using the input config:

```shell
$ kpt fn eval --image gcr.io/kpt-fn/create-setters:unstable --fn-config ./create-setters-fn-config.yaml
```

Alternatively, setter values can be passed as key-value pairs in the CLI

```shell
$ kpt fn eval --image gcr.io/kpt-fn/create-setters:unstable -- deploy=ubuntu-deployment env=ubuntu image=nginx tag=1.1.2
```

Modified resource looks like the following:

```yaml
# resources.yaml
apiVersion: v1
kind: Deployment
metadata:
  name: ubuntu-deployment-1 # kpt-set: ${deploy}-1
spec:
  image: ubuntu # kpt-set: ${env}
  app: "nginx:1.1.2" # kpt-set: ${image}:${tag}
  os:
    - ubuntu # kpt-set: ${env}
    - mac
```

>? This function doesn't add comments to scalar nodes with multi-line values.

Explanation for the changes:

`Comment` is added to the `Resource Field` value node when they match the `Scalar Setters`.

| Scalar Setters            | Resource Field                | Comment                            | Description     |
|---------------------------|---------------------------|------------------------------------|-----------------|
| <pre>env: ubuntu</pre>    | <pre>image: ubuntu</pre>  | `# kpt-set: ${env}`    | Setter value of `env` matches with value of `image`  |
| <pre>env: ubuntu</pre>    | <pre>  - ubuntu</pre>     | `# kpt-set: ${env}`    | Scalar value of the array node `os` matches with the setter value  of `env` |
| <pre>image: nginx<br>tag: 1.1.2</pre> | <pre>app: "nginx:1.1.2"</pre> | `# kpt-set: ${image}:${tag}`       | Non-overlapping substrings of resource field matches with the mentioned setter values. Resource field is parameterized with corresponding setter names.    |
| <pre>deploy: ubuntu-deployment</pre>  | <pre>name: ubuntu-deployment-1</pre> | `# kpt-set: ${deploy}-1`  | Overlapping substrings of resource value matches with the setter values of `deploy`  and `env`. Setter with longest length match `deploy` is considered.         |

### Setting comments for array nodes

Fields with array values can also be parameterized using setters. Since the values of `ConfigMap`
in the pipeline definition must be of string type, the array values must be wrapped into
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
role: [stage, dev]
```

Declare the array values, wrapped into string. Here the order of the array values
doesn't make a difference.

```yaml
# create-setters-fn-config.yaml
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
$ kpt fn eval --image gcr.io/kpt-fn/create-setters:unstable --fn-config ./create-setters-fn-config.yaml
```

Rendered resource looks like the following:

```yaml
# resources.yaml
apiVersion: v1
kind: MyKind
metadata:
  name: foo
environments: # kpt-set: ${env}
  - dev
  - stage
role: # kpt-set: ${env}
  - stage
  - dev
```

Explanation for the changes:
- As all the values in `environments` match the setter values of `env`, `# kpt-set: ${env}` comment is added.
Here, the comment is added to the key node as it is an array node with folded style.
- As all the values in `role` match the setter values of `env`, array node is converted to folded style and 
`# kpt-set: ${env}` comment is added to the key node. Here, the order of the array values is not considered.
<!--mdtogo-->

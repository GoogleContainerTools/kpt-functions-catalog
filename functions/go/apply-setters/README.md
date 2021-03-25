# apply-setters

### Overview

<!--mdtogo:Short-->

Apply setter values on resource fields. Setters serve as parameters for template-free
setting of field values.

<!--mdtogo-->

Setters are a safer alternative to other substitution techniques which do not have the context
of the structured data. Setters may be invoked to modify the package resources
using `apply-setters` function to set values.

### Synopsis

<!--mdtogo:Long-->

Package publishers declare setters in the package, consumers can set their values
either declaratively or imperatively.

Setter names can be discovered in the pipeline section of Kptfile, and the values
can be declared next to setter names.

```yaml
apiVersion: v1alpha2
kind: Kptfile
metadata:
  name: my-pkg
pipeline:
  mutators:
    - image: gcr.io/kpt-fns/apply-setters:unstable
      configMap:
        setter_name1: setter-value1
        setter_name2: setter-value2
```

The declared values for setters are rendered by invoking the following command:

```
kpt fn render [PKG_PATH]
```

Alternatively, this function can be invoked imperatively on the package by passing the
inputs as key-value pairs.

```
kpt fn eval gcr.io/kpt-fns/apply-setters:VERSION [PKG_PATH] -- [setter_name=setter_value]
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

Discover the names of setters in the Kptfile and declare desired values.

```yaml
apiVersion: v1alpha2
kind: Kptfile
metadata:
  name: my-pkg
pipeline:
  mutators:
    - image: gcr.io/kpt-fns/apply-setters:unstable
      configMap:
        image: ubuntu
        replicas: 3
```

Render the declared values by invoking:

```
kpt fn render
```

Alternatively, values can be rendered imperatively

```
kpt fn eval gcr.io/kpt-fns/apply-setters:unstable -- 'replicas=3'
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
apiVersion: v1alpha2
kind: Kptfile
metadata:
  name: my-pkg
pipeline:
  mutators:
    - image: gcr.io/kpt-fns/apply-setters:unstable
      configMap:
        env: |
          - prod
          - dev
```

Render the declared values by invoking:

```
kpt fn render
```

Alternatively, values can be rendered imperatively

```
kpt fn eval gcr.io/kpt-fns/apply-setters:unstable -- 'env=[prod, dev]'
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

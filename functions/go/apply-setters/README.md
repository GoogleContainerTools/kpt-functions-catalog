# apply-setters

### Overview

<!--mdtogo:Short-->

Apply setter values on resources fields. May set either the complete or partial field value.

<!--mdtogo-->

Setters provide a solution for template-free setting of field values. They are a
safer alternative to other substitution techniques which do not have the context
of the structured data. Setters may be invoked to modify the configuration
using `apply-setters` function to set values.

### Synopsis

<!--mdtogo:Long-->

```
kpt fn eval apply-setters:VERSION [DIR] -- [setter_name=setter_value]
```

Data model

1. Fields reference setters specified as line comments -- e.g.

```
    # kpt-set: replicas
```

2. Input values to setters are provided as key-value pairs -- e.g.

```
    kpt fn eval apply-setters:unstable -- replicas=3
```

Control flow

1. Read the package resources.
2. Locate all fields which reference the setter and change their values.
3. Write the modified resources back to the package.

<!--mdtogo-->

### Examples

<!--mdtogo:Examples-->

Let's start with the input resources

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: the-map # kpt-set: ${name}
data:
  some-key: some-value
---
apiVersion: v1
kind: MyKind
metadata:
  name: ns
environments: # kpt-set: ${env}
  - dev
  - stage
```

Invoke apply-setters function on the input resources

```
kpt fn eval apply-setters:unstable -- 'name=my-map' 'env=[prod, dev]'
```

The resources are transformed to

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-map # kpt-set: ${name}
data:
  some-key: some-value
---
apiVersion: v1
kind: MyKind
metadata:
  name: ns
environments: # kpt-set: ${env}
  - prod
  - dev
```

<!--mdtogo-->

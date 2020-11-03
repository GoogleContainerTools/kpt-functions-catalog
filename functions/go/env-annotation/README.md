# Env to annotation function

This is an example of implementing a function that adds annotations to resources based on the env variables value.
It can be used to in cases when it's necessary to set annotation value that isn't known at the moment of
writing the function configuration. In that case it's possible just to set the name of the env variable
in the configMap and left the value empty. It function finds the env variable with the value it sets
annotation value with that value.

## Function implementation

The function is implemented as an image (see Dockerfile).

## Function invocation

The function is invoked by authoring a [local Resource](local-resource)
with `metadata.annotations.[config.kubernetes.io/function]` and running:

    kpt fn run local-resource/

This exits non-zero if there is an error.

## Running the Example

Run the function with:

    datestamp="$(date)" kpt fn run local-resource/

The generated resources will appear in local-resource/ and will contain the current date in check-datestamp field:

```
$ cat local-resource/data.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
  annotations:
    datestamp: 'Tue Nov  3 05:05:47 UTC 2020'
data:
  value: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: map2
  annotations:
    datestamp: 'Tue Nov  3 05:05:47 UTC 2020'
data:
  value: value2
```

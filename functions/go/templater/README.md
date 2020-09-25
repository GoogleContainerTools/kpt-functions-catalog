# Templater function

This is an example of implementing a templater function.
Templater allows generating documents using go-template
engine and Sprig library.
`ConfigMap` may be used as a function configuration.
The function doesn't check GVKN of the configuration.
`data.template` field contains the template that will be
used by go-template engine.
If `data.cleanPipeline` field is true filter will remove
all previously exited documents in pipeline and will add
only generated. By default this function appends the 
pipeline.
All other literals inside `data` will be passed as values.

It's allowed to use not only scalars as values, but also
complex objects.

The idea is similar to helm, but in contrast with helm,
it doesn't require chart folder or link and allows
such Sprig-functions as env and expandenv (see [details](http://masterminds.github.io/sprig/os.html))
that gives more flexibility to build documents based
on combination of values written in the function
configuration and environment variables.

## Function implementation

The function is implemented as an image (see Dockerfile).

## Function invocation

The function is invoked by authoring a [local Resource](local-resource)
with `metadata.annotations.[config.kubernetes.io/function]` and running:

    kpt fn run local-resource/

This exits non-zero if there is an error.

## Running the Example

Run the function with:

    kpt fn run local-resource/

The generated resources will appear in local-resource/

```
$ cat local-resource/*

apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: node-1
spec:
  bootMACAddress: 00:aa:bb:cc:dd

apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: node-2
spec:
  bootMACAddress: 00:aa:bb:cc:ee
...
```

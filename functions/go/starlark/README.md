# starlark

## Overview

<!--mdtogo:Short-->

The `starlark` function contains a Starlark interpreter to run a Starlark script
to mutate or validate resources.

The starlark script can be used to:

- Add an annotation on the basis of a condition
- Inject a sidecar container in all KRM resources that contain a `PodTemplate`.
- Validate all KRM resources that contain a `PodTemplate` to ensure no
  containers run as root.

It follows the [executable configuration] pattern. It makes writing simple
function much easier.

<!--mdtogo-->

## Starlark Language

[Starlark] is a dialect of Python. It is commonly used as a configuration
language. It is an untyped dynamic language with high-level data types,
first-class functions with lexical scope, and garbage collection. You can find
the spec for the Starlark language [here][spec]. You can also find its API
reference [here][apiref].

<!--mdtogo:Long-->

## Usage

You need to put your starlark script source in the `functionConfig` of
kind `StarlarkRun` and then the function will run the starlark script that you
provide.

This function can be used both declaratively and imperatively.

### FunctionConfig

There are 2 kinds of `functionConfig` supported by this function:

- `ConfigMap`
- A custom resource of kind `StarlarkRun`

To use a `ConfigMap` as the `functionConfig`, the starlark script source must be
specified in the `data.source` field. Additional parameters can be specified in
the `data` field.

Here's an example:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: set-replicas
data:
  replicas: "5"
  source: |
    def setReplicas(resources, replicas):
      for r in resources:
        if r["apiVersion"] == "apps/v1" and r["kind"] == "Deployment":
          r["spec"]["replicas"] = replicas
    replicas = ctx.resource_list["functionConfig"]["data"]["replicas"]
    setReplicas(ctx.resource_list["items"], replicas)
```

In the example above, the script accesses the `replicas` parameters
using `ctx.resource_list["functionConfig"]["data"]["replicas"]`.

To use a `StarlarkRun` as the `functionConfig`, the starlark script source must
be specified in the `source` field. Additional parameters can be specified in
the `params` field. The `params` field supports any complex data structure as
long as it can be represented in yaml.

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: StarlarkRun
metadata:
  name: conditionally-add-annotations
params:
  toMatch:
    config.kubernetes.io/local-config: "true"
  toAdd:
    configmanagement.gke.io/managed: disabled
source: |    
  toMatch = ctx.resource_list["functionConfig"]["params"]["toMatch"]
  toAdd = ctx.resource_list["functionConfig"]["params"]["toAdd"]
  for resource in ctx.resource_list["items"]:
    match = True
    for key in toMatch:
      if key not in resource["metadata"]["annotations"] or resource["metadata"]["annotations"][key] != toMatch[key]:
        match = False
        break
    if match:
      for key in toAdd:
        resource["metadata"]["annotations"][key] = toAdd[key]
```

In the example above, the script accesses the `toMatch` parameters
using `ctx.resource_list["functionConfig"]["params"]["toMatch"]`.

There are 2 ways to run the function declaratively.

- Have your `Kptfile` with the inline `ConfigMap` as the `functionConfig`.
- Have your `Kptfile` pointing to a `functionConfig` file that contains either a
  `ConfigMap` or a `StarlarkRun`.

After that, you can render it with:

```shell
$ kpt fn render
```

There are 2 ways to run the function imperatively.

- Run it using a `ConfigMap` that generated from the command line arguments. The
  starlark script lives in `main.star` file.

```shell
$ kpt fn eval --image gcr.io/kpt-fn/starlark:v0.3.0 -- source="$(cat main.star)" param1=value1 param2=value2
```

- Run it using `--fn-config` with either a `ConfigMap` or a `StarlarkRun` that
  lives in `fn-config.yaml`.

```shell
$ kpt fn eval --image gcr.io/kpt-fn/starlark:v0.3.0 --fn-config fn-config.yaml
```

### Developing Starlark Script

Here's what you can do in the Starlark script:

- Read resources from `ctx.resource_list`. The `ctx.resource_list` complies with
  the [KRM Functions Specification]. You can read the input resources from
  `ctx.resource_list["items"]` and the `functionConfig` from
  `ctx.resource_list["functionConfig"]`.
- Write resources to `ctx.resource_list["items"]`.
- Read the environment variables. e.g. `ctx.environment["PATH"]`.
- Read the OpenAPI schema. e.g. `ctx.open_api["definitions"]["io.k8s.api.apps.v1.Deployment"]`
- Return an error using [`fail`][fail].
- Write error message to stderr using [`print`][print]

Here's what you currently cannot do in the Starlark script:

- We don't support the `results` field yet, i.e. if you write some results in
  `ctx.resource_list["results"]`, they will be ignored.
- While Starlark programs don't support working with yaml comments on resources,
  kpt will attempt to retain comments by copying them from the function inputs
  to the function outputs.

The starlark function has enabled the following non-standard starlark features:

- set: allow the `set` built-in. e.g. `s=set(["foo", "bar"])`
- recursion: allow while statements and recursive functions.
- global reassign: allow reassignment to top-level names; also, allow
  if/for/while at top-level.

In the standard Starlark, a [for loop] is permitted only within a function
definition. But in the starlark function, you can conveniently use `for`
statement at the top-level.

#### Libraries

We support the following [Starlib libraries]:

| Name               | How to load                            | Example |
|--------------------|----------------------------------------|---------|
| [bsoup]            | load('bsoup.star', 'bsoup')            | [example](https://github.com/qri-io/starlib/blob/master/bsoup/testdata/test.star)           |
| [encoding/base64]  | load('encoding/base64.star', 'base64') | [example](https://github.com/qri-io/starlib/blob/master/encoding/base64/testdata/test.star) |
| [encoding/csv]     | load('encoding/csv.star', 'csv')       | [example](https://github.com/qri-io/starlib/blob/master/encoding/csv/testdata/test.star)    |
| [encoding/json]    | load('encoding/json.star', 'json')     | [example](https://github.com/google/starlark-go/blob/master/starlark/testdata/json.star)    |
| [encoding/yaml]    | load('encoding/yaml.star', 'yaml')     | [example](https://github.com/qri-io/starlib/blob/master/encoding/yaml/testdata/test.star)   |
| [geo]              | load('geo.star', 'geo')                | [example](https://github.com/qri-io/starlib/blob/master/geo/testdata/test.star)             |
| [hash]             | load('hash.star', 'hash')              | [example](https://github.com/qri-io/starlib/blob/master/hash/testdata/test.star)            |
| [html]             | load('html.star', 'html')              | [example](https://github.com/qri-io/starlib/blob/master/html/testdata/test.star)            |
| [http]             | load('http.star', 'http')              | [example](https://github.com/qri-io/starlib/blob/master/http/testdata/test.star)            |
| [math]             | load('math.star', 'math')              | [example](https://github.com/google/starlark-go/blob/master/starlark/testdata/math.star)    |
| [re]               | load('re.star', 're')                  | [example](https://github.com/qri-io/starlib/blob/master/re/testdata/test.star)              |
| [time]             | load('time.star', 'time')              | [example](https://github.com/google/starlark-go/blob/master/starlark/testdata/time.star)    |
| [xlsx]             | load('xlsx.star', 'xlsx')              | [example](https://github.com/qri-io/starlib/blob/master/xlsx/testdata/test.star)            |
| [zipfile]          | load('zipfile.star', 'ZipFile')        | [example](https://github.com/qri-io/starlib/blob/master/zipfile/testdata/test.star)         |

### Debugging

It is possible to debug the `starlark` functions using [`print`][print].

For example, you can add something like the following in your Starlark script:

```python
print(ctx.resource_list["items"][0]["metadata"]["name"])
```

Then you can run the function:

```shell
kpt fn render --results-dir /tmp
```

You will find your debugging output in `functionResultList items.stderr`.

<!--mdtogo-->

[Starlark]: https://docs.bazel.build/versions/master/skylark/language.html

[executable configuration]: https://kpt.dev/book/05-developing-functions/04-executable-configuration

[spec]: https://github.com/bazelbuild/starlark/blob/master/spec.md

[apiref]: https://docs.bazel.build/versions/master/skylark/lib/skylark-overview.html

[KRM Functions Specification]: https://kpt.dev/book/05-developing-functions/01-functions-specification

[for loop]: https://github.com/bazelbuild/starlark/blob/master/spec.md#for-loops

[fail]: https://docs.bazel.build/versions/master/skylark/lib/globals.html#fail

[print]: https://docs.bazel.build/versions/master/skylark/lib/globals.html#print

[Starlib libraries]: https://github.com/qri-io/starlib#packages

[bsoup]: https://github.com/qri-io/starlib/tree/v0.5.0/bsoup

[encoding/base64]: https://github.com/qri-io/starlib/tree/v0.5.0/encoding/base64

[encoding/csv]: https://github.com/qri-io/starlib/tree/v0.5.0/encoding/csv

[encoding/json]: https://pkg.go.dev/go.starlark.net/lib/json

[encoding/yaml]: https://github.com/qri-io/starlib/tree/v0.5.0/encoding/yaml

[geo]: https://github.com/qri-io/starlib/tree/v0.5.0/geo

[hash]: https://github.com/qri-io/starlib/tree/v0.5.0/hash

[html]: https://github.com/qri-io/starlib/tree/v0.5.0/html

[http]: https://github.com/qri-io/starlib/tree/v0.5.0/http

[math]: https://pkg.go.dev/go.starlark.net/lib/math

[re]: https://github.com/qri-io/starlib/tree/v0.5.0/re

[time]: https://pkg.go.dev/go.starlark.net/lib/time

[xlsx]: https://github.com/qri-io/starlib/tree/v0.5.0/xlsx

[zipfile]: https://github.com/qri-io/starlib/tree/v0.5.0/zipfile

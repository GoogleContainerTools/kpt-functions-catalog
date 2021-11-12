# generate-blueprint-docs: Simple Example

### Overview

This example shows how the [generate-blueprint-docs] function works.

Running `generate-blueprint-docs` function on the example package will:

1. Generate a `GENERATED.md` file with the blueprint readme.

### Fetch the example package

Get the example package by running the following commands:

```shell
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/contrib/examples/generate-blueprint-docs-simple
```

### Function invocation

Invoke the function with the following command:

```shell
kpt fn eval -i gcr.io/kpt-fn-contrib/blueprint-docs:unstable --include-meta-resources \
--mount type=bind,src="$(pwd)",dst="/tmp",rw=true -- readme-path=/tmp/GENERATED.md
```

### Expected result

1. File `GENERATED.md` will be created with the generated readme.

[generate-blueprint-docs] https://catalog.kpt.dev/generate-blueprint-docs/v0.1/?id=definitions

Set Namespace
==================================================

Function `set-namespace` sets namespaces in KRM objects.

## Get Image

If you want to use the cutting-edge version of this function, you need to clone
the entire repo locally and then run the following to build the image:

<!-- @buildImage @testMaster -->
```
TAG=master
IMAGENAME=gcr.io/kpt-functions-trusted/set-namespace:$TAG
docker build -t $IMAGENAME ../../../functions/go/set-namespace/
```

If you want to use the latest released version:

<!-- @selectImage @testStable -->
```
TAG=v0.1
IMAGENAME=gcr.io/kpt-functions-trusted/set-namespace:$TAG
```

## Run Function

Let's first get the sample package using kpt:

```
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-namespace/helloworld/configs .
```

Next, let's create a directory for the results:

<!-- @createResults @testMaster @testStable -->
```
rm -rf results
mkdir results
```

Let's run the function [imperatively](https://googlecontainertools.github.io/kpt/guides/consumer/function/#imperative-run).
It will read the input objects from the `configs/` directory and then write the
output objects to the `results/` directory.

<!-- @runKptFunction @testMaster @testStable -->
```
kpt fn source configs/ | kpt fn run --image $IMAGENAME -- namespace=foo | kpt fn sink results/
```

If you want to write back to the same directory, you can simplify it to:
```
kpt fn run configs/ --image $IMAGENAME -- namespace=foo
```

Let's see what has been changed:

```
diff results/ configs/
```

If you have cloned the entire repo, we can check if we have gotten what we expect:

<!-- @compareWithGolden @testMaster @testStable -->
```
diff results/ golden/
```

## Cleanup

Let's remove the `results/` directory.

<!-- @cleanup @testMaster @testStable -->
```
rm -rf results
```

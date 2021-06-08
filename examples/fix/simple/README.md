# fix: Simple Example

In this example, we will fix a simple package which is compatible with v0.X version of kpt,
and make it compatible with kpt 1.0 

Let's start with the input resources

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: the-map # {"$kpt-set":"name"}
data:
  some-key: some-value
```

Here is an example Kptfile in the package:

```yaml
apiVersion: kpt.dev/v1alpha1
kind: Kptfile
metadata:
  name: nginx
packageMetadata:
  shortDescription: describe this package
upstream:
  type: git
  git:
    commit: 4d2aa98b45ddee4b5fa45fbca16f2ff887de9efb
    repo: https://github.com/GoogleContainerTools/kpt
    directory: package-examples/nginx
    ref: v0.2
openAPI:
  definitions:
    io.k8s.cli.setters.name:
      x-k8s-cli:
        setter:
          name: name
          value: the-map
```

Invoking `fix` function on the package transforms the resources as follows:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-new-map # kpt-set: ${name}
data:
  some-key: some-value
```

<!-- @skip -->
```yaml
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: nginx
info:
  description: describe this package
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:unstable
      configMap:
        name: the-map
upstream:
  type: git
  updateStrategy: resource-merge
  git:
    directory: package-examples/nginx
    ref: v0.2
    repo: https://github.com/GoogleContainerTools/kpt
upstreamLock:
  type: git
  git:
    directory: package-examples/nginx
    ref: v0.2
    repo: https://github.com/GoogleContainerTools/kpt
```

The transformed package is compatible with kpt 1.0 binary.

### Function invocation

Get the config example and try it out by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/fix/simple
$ kpt fn eval simple --image gcr.io/kpt-fn/fix:unstable --include-meta-resources
```

### Expected result

Check the resources in the package are transformed as described above.

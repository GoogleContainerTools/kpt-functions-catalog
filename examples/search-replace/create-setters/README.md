# search-replace: Create Setters Example

The `search-replace` function can be used to search for fields using available matchers
and add [setter] patterns as line comments in the resources.

This is an end to end example depicting [setter] creation process using `search-replace` function.

Let's start with the input resource

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-nginx
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: nginx
          image: "gcr.io/nginx:1.14.2"
```

Suppose you want to expose the values of `replicas`, `image` and `tag` as parameters.
You can create [setters] by invoking `search-replace` function with following arguments:

```sh
kpt fn run --image gcr.io/kpt-fn/search-replace:v0.1 'by-path=spec.replicas' 'put-comment=kpt-set: ${replicas}'
```

```sh
kpt fn run --image gcr.io/kpt-fn/search-replace:v0.1 'by-path=spec.**.image' 'put-comment=kpt-set: gcr.io/${image}:${tag}'
```

Transformed resource:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-nginx
spec:
  replicas: 3 # kpt-set: ${replicas}
  template:
    spec:
      containers:
        - name: nginx
          image: "gcr.io/nginx:1.14.2" # kpt-set: gcr.io/${image}:${tag}
```

Create `apply-setters` function config file and manually add created [setters] information to it.
This file can be used by package consumers to discover and pass new [setter] values.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: apply-setters-fn-config
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/kpt-fn/apply-setters:v0.1
data:
  # you may add description as comments
  replicas: 3
  image: nginx
  tag: 1.14.2
```

### Function invocation

Get the config example and try it out by running the following commands:

<!-- @getAndRunPkg @test -->
```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/search-replace/create-setters .
kpt fn run create-setters
```

### Expected result

Verify that the setter comments are added as depicted in the transformed resource above.

```sh
$ kpt cfg cat create-setters
```

Make sure that you add setters info to `apply-setters` function config as described above.

[setter]: https://catalog.kpt.dev/apply-setters/v0.1/
[setters]: https://catalog.kpt.dev/apply-setters/v0.1/
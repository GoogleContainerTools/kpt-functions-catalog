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

### Function invocation

Get the config example:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/search-replace/create-setters
```

Suppose you want to expose the values of `image` and `tag` as parameters.
You can create [setters] by invoking `search-replace` function with following arguments:

```shell
$ kpt fn eval --image gcr.io/kpt-fn/search-replace:unstable -- 'by-path=spec.replicas' 'put-comment=kpt-set: ${replicas}'
```

```shell
$ kpt fn eval --image gcr.io/kpt-fn/search-replace:unstable -- 'by-path=spec.**.image' 'put-comment=kpt-set: gcr.io/${image}:${tag}'
```

### Expected result

Verify that the setter comments are added as below:

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

Next, you can try to run the `apply-setters` function to use the [setters] that
you just created. For example:
```shell
$ kpt fn eval --image gcr.io/kpt-fn/search-replace:unstable -- replicas=3 image=nginx tag=1.14.2
```

You should be able to see the values got updated by the [setters].

[setter]: https://catalog.kpt.dev/apply-setters/v0.1/
[setters]: https://catalog.kpt.dev/apply-setters/v0.1/
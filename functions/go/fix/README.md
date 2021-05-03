# fix

### Overview

<!--mdtogo:Short-->

Fix resources and make them compatible with kpt 1.0.

<!--mdtogo-->

### Synopsis

<!--mdtogo:Long-->

`fix` helps you migrate the resources from `v1alpha1` format to `v1alpha2` format.
This is an automated step to migrate kpt packages which are compatible with kpt v0.X.Y
versions of kpt, and make them compatible with kpt 1.0

Here are the automated changes performed by `fix` function on `v1alpha1` kpt package:

1. The `packageMetaData` section will be transformed to `info` section.
2. `upstream` section(if present), in the `v1alpha1` Kptfile is converted to `upstream`
   and `upstreamLock` sections in `v1alpha2` version of Kptfile.
3. `dependencies` section will be removed from the Kptfile.
4. Setters no longer follow the openapi format. The setters and substitutions will be converted 
   to simple setter patterns. `apply-setters` function and transformed setters
   will be added to the mutators section in the `pipeline` section.
5. Function configs will be transformed(function annotation will be removed) and corresponding 
   definitions will be added to Kptfile.

<!--mdtogo-->

### Examples

<!--mdtogo:Examples-->

Let's start with a simple input resource which is compatible with kpt v0.X.Y

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
        - name: nginx # {"$kpt-set":"image"}
          image: nginx:1.14.1 # {"$kpt-set":"fullimage"}
          ports:
            - containerPort: 80
```

Here is the corresponding v1alpha1 Kptfile in the package

```yaml
apiVersion: kpt.dev/v1alpha1
kind: Kptfile
metadata:
  name: nginx
openAPI:
  definitions:
    io.k8s.cli.setters.image:
      x-k8s-cli:
        setter:
          name: image
          value: nginx
    io.k8s.cli.setters.tag:
      x-k8s-cli:
        setter:
          name: tag
          value: 1.14.1
    io.k8s.cli.substitutions.fullimage:
      x-k8s-cli:
        substitution:
          name: fullimage
          pattern: ${image}:${tag}
          values:
            - marker: ${image}
              ref: "#/definitions/io.k8s.cli.setters.image"
            - marker: ${tag}
              ref: "#/definitions/io.k8s.cli.setters.tag"
```

Invoke `fix` function on the package:

```sh
$ kpt fn eval --image gcr.io/kpt-fn/fix:unstable --include-meta-resources
```

Here is the transformed resource

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
        - name: nginx # kpt-set: ${image}
          image: nginx:1.14.1 # kpt-set: ${image}:${tag}
          ports:
            - containerPort: 80
```

Here is the transformed v1alpha2 Kptfile:

```yaml
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: nginx
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configMap:
        image: nginx
        tag: 1.14.1
```

<!--mdtogo-->

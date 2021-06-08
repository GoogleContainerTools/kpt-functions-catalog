# upsert-resource: Simple Example

In this example, we will see how `upsert-resource` function replaces the
matching resource (identified by GKNN (Group, Kind, Namespace and Name)) in the
package with the input resource.

Let's start with the list of resources in a package:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: myService
  namespace: mySpace
spec:
  selector:
    app: foo
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
  namespace: mySpace
spec:
  replicas: 3
```

Resource to upsert:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: myService
  namespace: mySpace
spec:
  selector:
    app: bar
```

Invoking `upsert-resource` function replaces the resource with name `myService`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: myService
  namespace: mySpace
spec:
  selector:
    app: bar
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
  namespace: mySpace
spec:
  replicas: 3
```

### Function invocation

Get the config example and try it out by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/upsert-resource/simple
$ kpt fn render simple
```

### Expected result

Check the resource with name `myService` is replaced with input resource. The
value of field `app` is updated.

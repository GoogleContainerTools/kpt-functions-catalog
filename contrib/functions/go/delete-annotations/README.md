# delete-annotations

## Overview

<!--mdtogo:Short-->

Deletes the supplied annotation keys from the resource(s) in a package. It can delete multiple annotations provided as a comma separated string as part of the function config (usage instructions below).

<!--mdtogo:Long-->

## Usage

The function will execute as follows:

1. Searches for resources with valid metadata
2. Deletes the annotation keys provided in KptFile or imperatively as config parameters

To execute imperatively:
```shell
$ kpt fn eval -i gcr.io/kpt-fn-contrib/delete-annotations:unstable -- annotationKeys=annotation.to.delete,another.annotation.to.delete
```

To execute `delete-annotations` declaratively include the function in kpt package pipeline as follows:
```yaml
...
pipeline:
  mutators:
    - image: gcr.io/kpt-fn-contrib/delete-annotations:unstable
      configMap:
        annotationKeys: annotation.to.delete,another.annotation.to.delete
...
```

<!--mdtogo-->

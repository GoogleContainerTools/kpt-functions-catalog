# Versioning

## SemVer and shorter SemVer

We use [semantic versioning] for all of our images. We also support a shorter
version of semantic versioning. E.g. `v<major> and v<major>.<minor>`.

The shorter version of semantic versioning is going to be "moving tags":

- `v<major>.<minor>` always points to `v<major>.<minor>.X` where X is the latest
(largest) patch version number. E.g. assume `v1.2` points `v1.2.0` initially.
After `v1.2.1` is released, `v1.2` now points to `v1.2.1`.
  
- `v<major>` always points to `v<major>.X` where X is the latest (largest) minor
  version number. E.g. assume `v1` points `v1.2.0` initially. After `v1.3.0` is
  released, `v1` now points to `v1.3.0`. After `v1.3.1` is released, `v1` now
  points to `v1.3.1`.

Note: We do NOT support the `latest` tag, since we cannot provide any
compatibility guarantee for it, and the pipeline won’t produce deterministic
results.

## API

All functions in the catalog comply with the [function spec]. A function can be
represented as the following:

![function representation](https://kpt.dev/static/images/func.svg)

There are 2 API surfaces for a function:
- The CRD function config API
- The behavior of a function as API

### CRD Function Config API

A function can choose to support a CRD as the function config. It can be
versioned as a normal CRD. 

### Function API

#### What are NOT part of the function API

- The formatting of serialization for output items and results. e.g. yaml
  indentation and order of fields in a map.
- The order of resources in the output items.
- The order of result items in the results.
- The content of the unstructured messages in the results.

#### What are part of the function API

- The supported fields in the ConfigMap as the function config.
- The supported versions of the CRD function configs.
- How the function behave given the input items and function config:
  - The reminder aspects of the output items that are not mentioned in the
    previous section.
  - The reminder aspects of the results that are not mentioned in the
    previous section.

### Breaking Changes

We define a breaking change as: For any given input (including input items,
functionConfig and OpenAPI), the function produces a different output (including
output items, results) that are part of the API.

### Backwards Compatibility

If there are breaking changes, we bump the major version.

If there are only bug fixes and security fixes (e.g. dependency package
non-breaking version bump and base image non-breaking version bump) in a
release, we bump the patch version.

In other cases (e.g. adding additional function parameters in a backward
compatible way), we bump the minor version.

Users won’t observe breaking changes if they are using the shorter semantic
versions (e.g. `v1.2` and `v1`) and they can automatically get the latest secure
patch version for free.

## Best Practices

- It is recommended to use shorter semantic versions (e.g. `v1.2` and `v1`) in
  your hydration pipeline.
  - Use `vX.Y` (e.g. `v0.1`) for pre-v1 functions that haven't reached a v1
    milestone.
  - Use `vX` (e.g. `v1`) for post-v1 functions.
- Don't use `latest` tag if you use your own function images, since it’s [not a
  best practice] for production and also it is [not recommended by kubernetes].

[not a best practice]: https://vsupalov.com/docker-latest-tag/
[not recommended by kubernetes]: https://kubernetes.io/docs/concepts/configuration/overview/#container-images
[semantic versioning]: https://semver.org/
[function spec]: https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md

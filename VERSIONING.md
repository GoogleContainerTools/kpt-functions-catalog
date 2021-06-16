# Versioning

## SemVer and shorter SemVer

We use [semantic versioning] for all of our images. The images with fully
specified version (e.g. `vX.Y.Z`) are immutable and will never be changed. We
also support a shorter version of semantic versioning. E.g.
`v<major>` and `v<major>.<minor>`.

The shorter version of semantic versioning is going to be floating tags:

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

## Interfaces

All functions in the catalog comply with the [function spec]. To learn more
about the functions concept,
see [here](http://kpt.dev/book/02-concepts/03-functions).

There are 2 interfaces for a function:

- The `functionConfig`
- The function itself

### functionConfig as an interface

A `functionConfig` can be either a core resource (e.g. `ConfigMap`) or a custom
resource. If the `functionConfig` is a CRD, it can be versioned independently as
a normal CRD.

### Function as an interface

#### What are NOT part of the function interface

- The formatting of serialization for output items and results. e.g. yaml
  indentation and order of fields in a map.
- The order of resources in the output items.
- The order of result items in the results.
- The content of the unstructured messages in the results.

#### What are part of the function interface

- The supported fields in the ConfigMap as the `functionConfig`.
- The supported versions of the CRD `functionConfigs`.
- How the function behave given the input items and `functionConfig`:
    - The reminder aspects of the output items that are not mentioned in the
      previous section.
    - The reminder aspects of the results that are not mentioned in the previous
      section.

### Breaking Changes

We define a breaking change as: For any given input (including `input items`,
`functionConfig` and OpenAPI), the function produces a different output
(including output items, results) that are part of the interface.

### Backwards Compatibility

For versions after v1.0.0, we will:

- Bump major version: There are breaking changes.

- Bump minor version: There are backward-compatible features.

- Bump patch version: There are only bug fixes and security fixes (e.g.
  dependency package non-breaking version bump and base image non-breaking
  version bump).

For versions before v1.0.0, the major version is always `0` and we will:

- Bumping minor version: There are breaking changes.

- Bumping patch version: In all other cases, including backward-compatible
  features, bug fixes and security fixes.

Users won’t observe breaking changes if they are using the shorter semantic
versions (e.g. `v1.2` and `v1`) and they can automatically get the latest secure
patch version for free.

## Best Practices

- It is recommended to use shorter semantic versions (e.g. `v1.2` and `v1`) in
  your hydration pipeline.
    - Use `vX.Y` (e.g. `v0.1`) for pre-v1 functions that haven't reached a v1
      milestone.
    - Use `vX` (e.g. `v1`) for post-v1 functions.
- Don't use `latest` tag if you use your own function images, since
  it’s [not a best practice] for production and also it
  is [not recommended by kubernetes].

[not a best practice]: https://vsupalov.com/docker-latest-tag/

[not recommended by kubernetes]: https://kubernetes.io/docs/concepts/configuration/overview/#container-images

[semantic versioning]: https://semver.org/

[function spec]: https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md

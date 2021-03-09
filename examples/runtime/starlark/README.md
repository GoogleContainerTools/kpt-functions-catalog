# Starlark Runtime

The `starlark` function provides a runtime to execute a starlark script that
updates or validate KRM resources.

Starlark source can be provided using `source` field.

- To provide it inline, use `.source.inline`.
- To provide it using a path, use `.source.path`.
- To provide it using a URL, use `.source.url`.

No matter how you provide starlark source, you can provide additional key-value
pairs using `keyValues` field. They can be accessed like this:

```
your_value = ctx.resource_list["functionConfig"]["keyValues"]["your_key"]
```

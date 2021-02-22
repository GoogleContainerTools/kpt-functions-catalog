# Set Name Prefix

The `set-name-prefix` KRM config function adds name prefix on
all resources except when users specify their own selectors.

## Function invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-name-prefix .
kpt fn run set-name-prefix --fn-path set-name-prefix
```

## Expected result

Check all resources have a name with prefix 'dev-':

```sh
kpt cfg cat set-name-prefix
```

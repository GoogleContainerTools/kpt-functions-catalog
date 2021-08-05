# sops: AGE example

### Overview

The `sops` KRM config function encrypts and decrypts resources. Learn more on
the [sops website].

This example demonstrates invocation of `sops` for encryption the resouce
called `toEncrypt` and decryption of the resource called `toDecrypt` using the
already existing AGE keys.

### Fetch the example package

Get this example and try it out by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/contrib/sops/decrypt-with-age .
```

There is a `age_keys.txt` file in the package that was downloaded
from https://raw.githubusercontent.com/mozilla/sops/master/age/keys.txt

## Function invocation

Invoke the function:

```shell
$ kpt fn eval -e SOPS_IMPORT_AGE="$(cat age_keys.txt)" -i gcr.io/kpt-fn-contrib/sops:unstable --fn-config fn-config.yaml
```

## Expected result

The resource named `toDecrypt` must be decrypted.

[sops website]: https://github.com/mozilla/sops#encrypting-using-age

# sops: PGP example

### Overview

The `sops` KRM config function encrypts and decrypts resources. Learn more info on the [sops website].

This example demonstrates invocation of `sops` for encryption the resouce called `toEncrypt` and decryption of the resource called `toDecrypt` using the already existing PGP keys.

## Function invocation

Get this example and try it out by running the following commands:

```shell
# download this example
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/contrib/sops/gpg .

# copy example GPG key from sops project
curl -fsSL -o gpg_keys.asc https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc

# run the function to work with GPG
SOPS_IMPORT_PGP="$(cat gpg_keys.asc)" kpt fn run gpg
```

## Expected result

Verify the updated configuration:

```shell
kpt cfg cat gpg
```

The resource called `toDecrypt` must be decrypted and the resource called `toEncrypt` must be encrypted.

[sops website]: https://github.com/mozilla/sops#test-with-the-dev-pgp-key

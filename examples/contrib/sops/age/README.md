# sops: AGE example

### Overview

The `sops` KRM config function encrypts and decrypts resources. Learn more on the [sops website].

This example demonstrates invocation of `sops` for encryption the resouce called `toEncrypt` and decryption of the resource called `toDecrypt` using the already existing AGE keys.

## Function invocation

Get this example and try it out by running the following commands:

```shell
# download this example
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/contrib/sops/age .

# copy example AGE key from sops project
curl -fsSL -o age_keys.txt https://raw.githubusercontent.com/mozilla/sops/master/age/keys.txt

# run the function to work with AGE
SOPS_IMPORT_AGE="$(cat age_keys.txt)" kpt fn run age
```

## Expected result

Verify the updated configuration:

```shell
kpt cfg cat age
```

The resource called `toDecrypt` must be decrypted and the resource called `toEncrypt` must be encrypted.

[sops website]: https://github.com/mozilla/sops#encrypting-using-age

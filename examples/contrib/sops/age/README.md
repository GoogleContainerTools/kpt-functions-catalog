# SOPS AGE example

SOPS function introduces for AGE case an additional `SOPS_IMPORT_AGE` ENV variable
that must contain the [SOPS age keys.txt file](https://github.com/mozilla/sops/blob/master/age/keys.txt).
This file is needed for decryption. For encryption it's possible to use
`SOPS_AGE_RECIPIENTS` ENV variable or parameter `age`.

E.g. for decryption it's possible to run:

```sh
kpt fn run --env SOPS_IMPORT_AGE="$(cat <file with age keys>.txt)" <folder>
```

or if AGE keys are already stored in the host system so SOPS binary works locally, it's possible to run:

```sh
kpt fn run --env SOPS_IMPORT_AGE="$(cat ~/.config/sops/age/keys.txt)" <folder>
```

## Function invocation

Get this example and try it out by running the following commands:

```sh
# download sops kpt-function example
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/contrib/sops/age .

# copy example AGE key from sops project
curl -fsSL -o age_keys.txt https://raw.githubusercontent.com/mozilla/sops/master/age/keys.txt

# run the function to work with AGE
kpt fn run --env SOPS_IMPORT_AGE="$(cat age_keys.txt)" age
```

## Expected result

Verify the updated configuration:

```sh
kpt cfg cat age
```

# SOPS PGP example

SOPS function introduces for PGP case an additional `SOPS_IMPORT_PGP` ENV variable
that must contain the private key(s) needed to decrypt yamls and public key(s) to
encrypt yamls. If you have a file with keys it's possible to run:

```sh
kpt fn run --env SOPS_IMPORT_PGP="$(cat <file with exported key>.asc)" <folder>
```

or if your keys are already in stored in `gpg`-storage, it's possible to run:

```sh
kpt fn run --env SOPS_IMPORT_PGP="$(gpg --armor --export-secret-keys)" <folder>
```

## Function invocation

Get this example and try it out by running the following commands:

```sh
# download sops kpt-function example
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/contrib/sops/gpg .

# copy example GPG key from sops project
curl -fsSL -o gpg_keys.asc https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc

# run the function to work with GPG
kpt fn run --env SOPS_IMPORT_PGP="$(cat gpg_keys.asc)" gpg
```

## Expected result

Verify the updated configuration:

```sh
kpt cfg cat gpg
```

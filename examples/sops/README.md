# SOPS

The `sops` config function transforms an input kpt package according to the
function configuration: in the current version it can decrypt the documents that
have `sops` field with SOPS metadata. This example invokes the `sops -d`
function using declarative configuration. See
[sops readme](https://github.com/mozilla/sops/blob/master/README.rst) for more
details.

In order to decrypt yaml, `sops` may accept a variety of ENV vars, e.g. to work
with Hashicorp Vault it will be necessary to set: `VAULT_ADDR` and
`VAULT_TOKEN`. For PGP case this function introduces `SOPS_IMPORT_PGP` ENV var
that must contain the private key or keys needed to decrypt yamls. If you have a
file with keys it's possible to run:

```sh
kpt fn run --env SOPS_IMPORT_PGP="$(cat <file with exported key>.asc)" <folder>
```

or if your keys are already in `gpg`, it's possibe to run:

```sh
kpt fn run --env SOPS_IMPORT_PGP="$(gpg --armor --export-secret-keys)" <folder>
```

## Function invocation

Get this example and try it out by running the following commands:

```sh
# copy example key from sops project
curl -fsSL -o sops/key.asc https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc
# download sops kpt-function example
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/sops .
# run the function
kpt fn run --env SOPS_IMPORT_PGP="$(cat sops/key.asc)" sops/local-configs
```

## Expected result

Verify the updated configuration:

```sh
kpt cfg cat sops/local-configs
```

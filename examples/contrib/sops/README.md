# SOPS

The `sops` config function transforms an input kpt package according to the
function configuration: in the current version it can encrypt yaml documents or
decrypt the yaml documents that have `sops` field with SOPS metadata. What exactly
operation will be performed is set by `cmd` field of the function configuration.
Depending on that field value the function invokes `sops -d` for decryption and 
`sops -e` for encryption. The function passes all other parameters directly
to the command-line parameters of `sops` tool: e.g. if it's necessary to pass
`--unencrypted-regex <value>`, it's possible just to add a field
`unencrypted-regex: <value>` to the function config. 
See [sops readme](https://github.com/mozilla/sops/blob/master/README.rst) for more
details on the parameters and environment variables that `sops` tool accepts.

There is an option to specify only a particular subset of input yaml documents that have to be
encrypted or decrypted: `cmd-json-path-filter` field accepts JSONPath notation to do so.
E.g. `cmd-json-path-filter: '$[?(@.metadata.name=="somename" &&@.kind=="somekind")]'` will process
only documents with name 'somename' and kind 'somekind'.

Another special option is `cmd-tolerate-failures: true` that ignores the `sops` tool errors on
any operation and just keeps the document unmodified in that case.

In order to encrypt or decrypt yaml, `sops` may accept a variety of ENV variables, e.g. to work
with Hashicorp Vault it will be necessary to set: `VAULT_ADDR` and
`VAULT_TOKEN`. This option can be used to set different encryption parameters that shouldn't be stored
in git repository, e.g. private keys, external services credentials.
This function introduces 2 additional ENV variables: `SOPS_IMPORT_PGP` and `SOPS_IMPORT_AGE` that
make possible to work with PGP and AGE encryption. Please refer to [gpg](gpg/) and [age](age/) examples
to get more details.

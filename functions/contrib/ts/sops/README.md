# SOPS

### Overview

Use [sops] to encrypt/decrypt KRM resources.

### FunctionConfig

The function configuration must be a ConfigMap.

The following keys can be used in the `data` field of the ConfigMap, and all of
them are optional:

- `cmd`: The operation that sops perform: `encrypt` or `decrypt`. The default
  is `decrypt`.
- `cmd-json-path-filter`: Operation will be performed only to the resources that match
  to the filter, e.g. `$[?(@.metadata.name=="somename" && @.kind=="somekind")]` will
  process only resources with name `somename` and kine `somekind`. The default is empty.
- `cmd-tolerate-failures`: Ignore sops error and keep KRM resource unchanged rather than
  exit the function with error. The default is `false`.
- `verbose`: Sops verbose logging output is enabled. The default is `false`.
- `ignore-mac`: Sops will ignore Message Authentication Code during decryption. The default
  is `false`.
- `override-preexec-cmd`: The command that will be executed prior to sops exectution. The default is
  `[ "$SOPS_IMPORT_PGP" == "" ] || (echo "$SOPS_IMPORT_PGP" | gpg --import 2>/dev/null); [ "$XDG_CONFIG_HOME" == "" ] || [ "$SOPS_IMPORT_AGE" == "" ] || (mkdir -p $XDG_CONFIG_HOME/sops/age/ && echo "$SOPS_IMPORT_AGE" > $XDG_CONFIG_HOME/sops/age/keys.txt`.
  This command allows to import encryption keys via ENV variables.
- `override-detached-annotations`: List of comma-separated annotations that will be removed from the KRM resource
  if exist prior to sops execution and added back after execution. That helps to avoid decryption issues
  in cases the composer (e.g. kpt) adds its internal annotations. The default is
  `config.kubernetes.io/index,config.kubernetes.io/path,config.k8s.io/id,kustomize.config.k8s.io/id`.
- all other provided keys will be converted to the sops command arguments using pattern `--<key name> <value>`, e.g.
  `unencrypted-regex: '^(kind|apiVersion|group|metadata)$'` will add sops parameter `--unencrypted-regex: '^(kind|apiVersion|group|metadata)$'`.

In order to encrypt or decrypt yaml, `sops` may accept a variety of ENV variables, e.g. to work
with Hashicorp Vault it will be necessary to set: `VAULT_ADDR` and
`VAULT_TOKEN`. This option can be used to set different encryption parameters that shouldn't be stored
in version control system repository, e.g. private keys, external services credentials.
This function introduces 2 additional ENV variables: `SOPS_IMPORT_PGP` and `SOPS_IMPORT_AGE` that must contain the PGP or AGE keys and that
make possible to work with PGP and AGE encryption.

For `pgp` if you have a file with keys it's possible to run:

```shell
SOPS_IMPORT_PGP="$(cat <file with exported key>.asc)" kpt fn run <folder>
```

or if your keys are already presented in `gpg`-storage, it's possible to run:

```shell
SOPS_IMPORT_PGP="$(gpg --armor --export-secret-keys)" kpt fn run <folder>
```

To make `sops` decrypt `age` it's necessary to keep all keys in the single file `~/.config/sops/age/keys.txt`. If that file exists, it's possible to invoke sops function and provide it with that keys by the command:

```shell
SOPS_IMPORT_AGE="$(cat ~/.config/sops/age/keys.txt)" kpt fn run <folder>
```

Please refer to [gpg] and [age] examples to get more details.

[age]:/examples/contrib/sops/age/
[gpg]:/examples/contrib/sops/gpg/
[sops]:https://github.com/mozilla/sops

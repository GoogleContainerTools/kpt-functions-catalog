# KPT Functions catalog

This repository contains a catalog of KPT functions.

| Image                                     | Description                                                                                                                | Use Case       |
| ----------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- | -------------- |
| gcr.io/kpt-functions/read-yaml            | Reads a directory of Kubernetes configuration recursively.                                                                 | Source         |
| gcr.io/kpt-functions/write-yaml           | Writes a directory of Kubernetes configuration. It maintains the original directory structure as read by source functions. | Sink           |
| gcr.io/kpt-functions/gatekeeper-validate  | Enforces OPA constraints on input objects. The constraints are also passed as part of the input to the function.           | Compliance     |
| gcr.io/kpt-functions/mutate-psp           | [Demo] Mutates `PodSecurityPolicy` objects by setting `spec.allowPrivilegeEscalation` to `false`.                          | Recommendation |
| gcr.io/kpt-functions/validate-rolebinding | [Demo] Enforces a blacklist of `subjects` in `RoleBinding` objects.                                                        | Compliance     |
| gcr.io/kpt-functions/expand-team-cr       | [Demo] Reads custom resources of type `Team` and generates multiple `Namespace` and `RoleBinding` objects.                 | Generation     |
| gcr.io/kpt-functions/no-op                | [Demo] No Op function.                                                                                                     | Testing        |

# Running functions

Functions can be piped to form sophisticated pipelines, for example:

```sh
git clone git@github.com:frankfarzan/foo-corp-configs.git
cd foo-corp-configs

docker pull gcr.io/kpt-functions/read-yaml
docker pull gcr.io/kpt-functions/mutate-psp
docker pull gcr.io/kpt-functions/expand-team-cr
docker pull gcr.io/kpt-functions/validate-rolebinding
docker pull gcr.io/kpt-functions/write-yaml

docker run -i -u $(id -u) -v $(pwd):/source  gcr.io/kpt-functions/read-yaml -i /dev/null -d source_dir=/source |
docker run -i gcr.io/kpt-functions/mutate-psp |
docker run -i gcr.io/kpt-functions/expand-team-cr |
docker run -i gcr.io/kpt-functions/validate-rolebinding -d subject_name=alice@foo-corp.com |
docker run -i -u $(id -u) -v $(pwd):/sink gcr.io/kpt-functions/write-yaml -o /dev/null -d sink_dir=/sink -d overwrite=true
```

Let's walk through each step:

1. Clone the `foo-corp-configs` repo containing example configs.
1. Pull all the docker images.
1. `read-yaml` function recursively **reads** all YAML files from `foo-corp-configs` directory on the host.
   It outputs the content of the directory in a standard format to `stdout`. By default, docker containers
   runs as a non-privileged user. You need to specify `-u` with your user id to access host files as shown above.
1. `mutate-psp` function reads the output of `read-yaml`. This function **mutates** any `PodSecurityPolicy` resources by setting a field called `allowPrivilegeEscalation` to `false`.
1. `expand-team-cr` function similarly operates on the result of the previous function. It looks
   for Kubernetes custom resource of kind `Team`, and based on that **generates** new resources (e.g. `Namespaces` and `RoleBindings`).
1. `validate-rolebinding` function **enforces** a policy that disallows any `RoleBindings` with `subject`
   set to `alice@foo-corp.com`. This steps fails with a non-zero exit code if this policy is violated.
1. `write-yaml` **writes** the result of the pipeline back to `foo-corp-configs` directory on the host.

Let's see what changes were made to the repo:

```sh
git status
```

You should see these changes:

1. `podsecuritypolicy_psp.yaml` should have been mutated by `mutate-psp` function.
1. `payments-dev` and `payments-prod` directories created by `expand-team-cr` function.

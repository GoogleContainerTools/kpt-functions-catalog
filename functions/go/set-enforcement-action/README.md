# set-enforcement-action

## Overview

<!--mdtogo:Short-->

Applies the supplied enforcement action on [policy constraints](https://cloud.google.com/anthos-config-management/docs/concepts/policy-controller#constraints) within a package.

<!--mdtogo-->

Policy Controller [constraint objects](https://cloud.google.com/anthos-config-management/docs/how-to/creating-constraints) enable you to enforce policies for your Kubernetes clusters.
This function provides a quick way for users to set the [enforcementAction](https://cloud.google.com/anthos-config-management/docs/how-to/auditing-constraints#types_of_enforcement_actions)
attribute to either `dryrun` for auditing, `warn` for letting the non-compliant resource be applied to the cluster with a warning or `deny` for enforcing the constraints and denying the resource application altogether.

<!--mdtogo:Long-->

## Usage

The function will execute as follows:

1. Searches for resources with `apiVersion: constraints.gatekeeper.sh/v1beta1`
2. Applies the enforement action value provided in KptFile to following element:
   `spec.enforcementAction`

`set-enforcement-action` function can be executed imperatively as follows:

```shell
$ kpt fn eval -i gcr.io/kpt-fn/set-enforcement-action:unstable -- enforcementAction=deny
```

To execute `set-enforcement-action` declaratively include the function in kpt package pipeline as follows:
```yaml
...
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-enforcement-action:unstable
      configMap:
          enforcementAction: deny
...
```

<!--mdtogo-->

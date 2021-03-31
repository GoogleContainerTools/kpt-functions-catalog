# gatekeeper-validate

### Overview

<!--mdtogo:Short-->

Validate the KRM resources using the policy controller.

<!--mdtogo-->

### Synopsis

<!--mdtogo:Long-->

You can use the policy controller to validate KRM resources. To learn more about
the policy controller, see: https://cloud.google.com/anthos-config-management/docs/concepts/policy-controller.

The function ensures the constraint policies are enforced on KRM resources.
The function takes 3 types of resources from the input resource list:

- constraint templates
- constraints
- other KRM resources

Every constraint should be backed by a constraint template that defines the
schema and logic of the constraint.

To learn more about how to write constraint templates and constraints, see:
https://cloud.google.com/anthos-config-management/docs/how-to/write-a-constraint-template
and
https://cloud.google.com/anthos-config-management/docs/how-to/creating-constraints.

<!--mdtogo-->

### Examples

<!--mdtogo:Examples-->

TODO: link to the examples

<!--mdtogo-->

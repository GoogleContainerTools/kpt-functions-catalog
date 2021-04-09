# gatekeeper-validate

### Overview

<!--mdtogo:Short-->

Validate the KRM resources using [Gatekeeper] constraints.

<!--mdtogo-->

### Synopsis

<!--mdtogo:Long-->

You can use Gatekeeper to validate KRM resources. To learn more about how to use
the Gatekeeper project, see: https://open-policy-agent.github.io/gatekeeper/website/docs/howto.

The function evaluates constraint policies against KRM resources.
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

<!-- TODO: update the following link to web page -->

<!--mdtogo:Examples-->

https://github.com/GoogleContainerTools/kpt-functions-catalog/tree/master/examples/gatekeeper-validate/

<!--mdtogo-->

[Gatekeeper]:https://github.com/open-policy-agent/gatekeeper
# gatekeeper

### Overview

<!--mdtogo:Short-->

Validate the KRM resources using [Gatekeeper] constraints.

<!--mdtogo-->

### FunctionConfig

<!--mdtogo:Long-->

[Gatekeeper] allows users to validate the KRM resources against the Gatekeeper
constraints.

You will need to define a [Constraint Template] first before defining a
[Constraint]. Every constraint should be backed by a constraint template that
defines the schema and logic of the constraint.
To learn more about how to use the Gatekeeper project, see [here].

At least one constraint template and at least one constraint must be provided
using `resourceList.items` along with other KRM resources. No function config is
needed in `resourceList.functionConfig`.

<!--mdtogo-->

[Gatekeeper]: https://open-policy-agent.github.io/gatekeeper/website/docs/
[Constraint Template]: https://open-policy-agent.github.io/gatekeeper/website/docs/howto#constraint-templates
[Constraint]: https://open-policy-agent.github.io/gatekeeper/website/docs/howto#constraints
[here]: https://open-policy-agent.github.io/gatekeeper/website/docs/howto
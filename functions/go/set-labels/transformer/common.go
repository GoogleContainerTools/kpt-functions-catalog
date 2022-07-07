package transformer

const (
	fnConfigGroup      = "fn.kpt.dev"
	fnConfigAPIVersion = "v1alpha1"
	fnConfigKind       = "SetLabels"
	fnDeprecateField   = "additionalLabelFields"
)

// generate common label paths
var CommonSpecs = FieldSpecs{
	FieldSpec{
		Identifier:         GVK{"", "v1", "Service"},
		Path:               FieldPath{"spec", "selector"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"", "v1", "ReplicationController"},
		Path:               FieldPath{"spec", "selector"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"", "v1", "ReplicationController"},
		Path:               FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"", "", "Deployment"},
		Path:               FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"", "", "Deployment"},
		Path:               FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "Deployment"},
		Path:               FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "preferredDuringSchedulingIgnoredDuringExecution", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "Deployment"},
		Path:               FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "requiredDuringSchedulingIgnoredDuringExecution", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "Deployment"},
		Path:               FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "preferredDuringSchedulingIgnoredDuringExecution", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "Deployment"},
		Path:               FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "requiredDuringSchedulingIgnoredDuringExecution", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "Deployment"},
		Path:               FieldPath{"spec", "template", "spec", "topologySpreadConstraints", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"", "", "ReplicaSet"},
		Path:               FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"", "", "ReplicaSet"},
		Path:               FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"", "", "DaemonSet"},
		Path:               FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"", "", "DaemonSet"},
		Path:               FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "StatefulSet"},
		Path:               FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "StatefulSet"},
		Path:               FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "StatefulSet"},
		Path:               FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "preferredDuringSchedulingIgnoredDuringExecution", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "StatefulSet"},
		Path:               FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "requiredDuringSchedulingIgnoredDuringExecution", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "StatefulSet"},
		Path:               FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "preferredDuringSchedulingIgnoredDuringExecution", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "StatefulSet"},
		Path:               FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "requiredDuringSchedulingIgnoredDuringExecution", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "StatefulSet"},
		Path:               FieldPath{"spec", "template", "spec", "topologySpreadConstraints", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"batch", "", "Job"},
		Path:               FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"batch", "", "Job"},
		Path:               FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"batch", "", "CronJob"},
		Path:               FieldPath{"spec", "jobTemplate", "spec", "selector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"batch", "", "CronJob"},
		Path:               FieldPath{"spec", "jobTemplate", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"batch", "", "CronJob"},
		Path:               FieldPath{"spec", "jobTemplate", "spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"policy", "", "PodDisruptionBudget"},
		Path:               FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"networking.k8s.io", "", "NetworkPolicy"},
		Path:               FieldPath{"spec", "podSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"networking.k8s.io", "", "NetworkPolicy"},
		Path:               FieldPath{"spec", "ingress", "from", "podSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"networking.k8s.io", "", "NetworkPolicy"},
		Path:               FieldPath{"spec", "egress", "to", "podSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
}

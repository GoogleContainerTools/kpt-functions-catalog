// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		Gvk:                GVK{"", "v1", "Service"},
		FieldPath:          FieldPath{"spec", "selector"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Gvk:                GVK{"", "v1", "ReplicationController"},
		FieldPath:          FieldPath{"spec", "selector"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Gvk:                GVK{"", "v1", "ReplicationController"},
		FieldPath:          FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Gvk:                GVK{"", "v1", "ReplicationController"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "preferredDuringSchedulingIgnoredDuringExecution[]", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"", "v1", "ReplicationController"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "requiredDuringSchedulingIgnoredDuringExecution[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "ReplicationController"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "preferredDuringSchedulingIgnoredDuringExecution[]", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "ReplicationController"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "requiredDuringSchedulingIgnoredDuringExecution[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "Deployment"},
		FieldPath:          FieldPath{"spec", "template", "spec", "topologySpreadConstraints[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"", "", "Deployment"},
		FieldPath:          FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Gvk:                GVK{"", "", "Deployment"},
		FieldPath:          FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "Deployment"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "preferredDuringSchedulingIgnoredDuringExecution[]", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "Deployment"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "requiredDuringSchedulingIgnoredDuringExecution[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "Deployment"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "preferredDuringSchedulingIgnoredDuringExecution[]", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "Deployment"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "requiredDuringSchedulingIgnoredDuringExecution[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "Deployment"},
		FieldPath:          FieldPath{"spec", "template", "spec", "topologySpreadConstraints[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"", "", "ReplicaSet"},
		FieldPath:          FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Gvk:                GVK{"", "", "ReplicaSet"},
		FieldPath:          FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Gvk:                GVK{"", "v1", "ReplicaSet"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "preferredDuringSchedulingIgnoredDuringExecution[]", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"", "v1", "ReplicaSet"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "requiredDuringSchedulingIgnoredDuringExecution[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "ReplicaSet"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "preferredDuringSchedulingIgnoredDuringExecution[]", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "ReplicaSet"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "requiredDuringSchedulingIgnoredDuringExecution[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "ReplicaSet"},
		FieldPath:          FieldPath{"spec", "template", "spec", "topologySpreadConstraints[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"", "", "DaemonSet"},
		FieldPath:          FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Gvk:                GVK{"", "", "DaemonSet"},
		FieldPath:          FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Gvk:                GVK{"", "v1", "DaemonSet"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "preferredDuringSchedulingIgnoredDuringExecution[]", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"", "v1", "DaemonSet"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "requiredDuringSchedulingIgnoredDuringExecution[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "DaemonSet"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "preferredDuringSchedulingIgnoredDuringExecution[]", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "DaemonSet"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "requiredDuringSchedulingIgnoredDuringExecution[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "DaemonSet"},
		FieldPath:          FieldPath{"spec", "template", "spec", "topologySpreadConstraints[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "StatefulSet"},
		FieldPath:          FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "StatefulSet"},
		FieldPath:          FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "StatefulSet"},
		FieldPath:          FieldPath{"spec", "volumeClaimTemplates[]", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "StatefulSet"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "preferredDuringSchedulingIgnoredDuringExecution[]", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "StatefulSet"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "requiredDuringSchedulingIgnoredDuringExecution[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "StatefulSet"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "preferredDuringSchedulingIgnoredDuringExecution[]", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "StatefulSet"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "requiredDuringSchedulingIgnoredDuringExecution[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "StatefulSet"},
		FieldPath:          FieldPath{"spec", "template", "spec", "topologySpreadConstraints", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"batch", "", "Job"},
		FieldPath:          FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"batch", "", "Job"},
		FieldPath:          FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Gvk:                GVK{"", "v1", "Job"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "preferredDuringSchedulingIgnoredDuringExecution[]", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"", "v1", "Job"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "requiredDuringSchedulingIgnoredDuringExecution[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "Job"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "preferredDuringSchedulingIgnoredDuringExecution[]", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "Job"},
		FieldPath:          FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "requiredDuringSchedulingIgnoredDuringExecution[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "Job"},
		FieldPath:          FieldPath{"spec", "template", "spec", "topologySpreadConstraints[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"batch", "", "CronJob"},
		FieldPath:          FieldPath{"spec", "jobTemplate", "spec", "selector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"batch", "", "CronJob"},
		FieldPath:          FieldPath{"spec", "jobTemplate", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Gvk:                GVK{"batch", "", "CronJob"},
		FieldPath:          FieldPath{"spec", "jobTemplate", "spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Gvk:                GVK{"batch", "", "CronJob"},
		FieldPath:          FieldPath{"spec", "jobTemplate", "spec", "template", "spec", "affinity", "podAffinity", "preferredDuringSchedulingIgnoredDuringExecution[]", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Gvk:                GVK{"", "v1", "CronJob"},
		FieldPath:          FieldPath{"spec", "jobTemplate", "spec", "template", "spec", "affinity", "podAffinity", "requiredDuringSchedulingIgnoredDuringExecution[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "CronJob"},
		FieldPath:          FieldPath{"spec", "jobTemplate", "spec", "template", "spec", "affinity", "podAntiAffinity", "preferredDuringSchedulingIgnoredDuringExecution[]", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "CronJob"},
		FieldPath:          FieldPath{"spec", "jobTemplate", "spec", "template", "spec", "affinity", "podAntiAffinity", "requiredDuringSchedulingIgnoredDuringExecution[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"apps", "", "CronJob"},
		FieldPath:          FieldPath{"spec", "jobTemplate", "spec", "template", "spec", "topologySpreadConstraints[]", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"policy", "", "PodDisruptionBudget"},
		FieldPath:          FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"networking.k8s.io", "", "NetworkPolicy"},
		FieldPath:          FieldPath{"spec", "podSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"networking.k8s.io", "", "NetworkPolicy"},
		FieldPath:          FieldPath{"spec", "ingress[]", "from[]", "podSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"networking.k8s.io", "", "NetworkPolicy"},
		FieldPath:          FieldPath{"spec", "egress[]", "to[]", "podSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
}

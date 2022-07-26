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
	FnConfigKind     = "SetLabels"
	fnDeprecateField = "additionalLabelFields"
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
		Gvk:                GVK{"policy", "", "PodDisruptionBudget"},
		FieldPath:          FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Gvk:                GVK{"networking.k8s.io", "", "NetworkPolicy"},
		FieldPath:          FieldPath{"spec", "podSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
}

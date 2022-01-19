package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func runNamespaceTransformerE(config, input string) (string, error) {
	resmapFactory := newResMapFactory()
	resMap, err := resmapFactory.NewResMapFromBytes([]byte(input))
	if err != nil {
		return "", err
	}

	var plugin *plugin = &KustomizePlugin
	err = plugin.Config(nil, []byte(config))
	if err != nil {
		return "", err
	}
	defaultConfig, err := getDefaultConfig()
	if err != nil {
		return "", err
	}
	if len(plugin.AdditionalNamespaceFields) == 0 {
		plugin.AdditionalNamespaceFields = defaultConfig.FieldSpecs
	}
	err = plugin.Transform(resMap)
	if err != nil {
		return "", err
	}
	y, err := resMap.AsYaml()
	if err != nil {
		return "", err
	}
	return string(y), nil
}

func runNamespaceTransformer(t *testing.T, config, input string) string {
	s, err := runNamespaceTransformerE(config, input)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestNamespaceTransformer1(t *testing.T) {
	config := `
namespace: test
`
	input := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm2
  namespace: foo
---
apiVersion: v1
kind: Service
metadata:
  name: svc1
---
apiVersion: v1
kind: Namespace
metadata:
  name: ns1
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: default
  namespace: test
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: service-account
  namespace: system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: manager-rolebinding
subjects:
- kind: ServiceAccount
  name: default
  namespace: system
- kind: ServiceAccount
  name: service-account
  namespace: system
- kind: ServiceAccount
  name: another
  namespace: random
---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: example
webhooks:
  - name: example1
    clientConfig:
      service:
        name: svc1
        namespace: system
  - name: example2
    clientConfig:
      service:
        name: svc2
        namespace: system
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: crd
`
	expected := `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1
  namespace: test
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm2
  namespace: test
---
apiVersion: v1
kind: Service
metadata:
  name: svc1
  namespace: test
---
apiVersion: v1
kind: Namespace
metadata:
  name: test
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: default
  namespace: test
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: service-account
  namespace: test
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: manager-rolebinding
subjects:
- kind: ServiceAccount
  name: default
  namespace: test
- kind: ServiceAccount
  name: service-account
  namespace: test
- kind: ServiceAccount
  name: another
  namespace: random
---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: example
webhooks:
- clientConfig:
    service:
      name: svc1
      namespace: system
  name: example1
- clientConfig:
    service:
      name: svc2
      namespace: system
  name: example2
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: crd
`

	output := runNamespaceTransformer(t, config, input)
	if output != expected {
		fmt.Println("Actual:")
		fmt.Println(output)
		fmt.Println("===")
		fmt.Println("Expected:")
		fmt.Println(expected)
		t.Fatalf("Actual doesn't equal to expected")
	}
}

func TestNamespaceTransformerClusterLevelKinds(t *testing.T) {
	input := `apiVersion: v1
kind: Namespace
metadata:
  name: ns1
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crd1
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cr1
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crb1
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv1
`

	config := `
namespace: test
fieldSpecs:
- path: metadata/namespace
  create: true
- path: subjects
  kind: RoleBinding
  group: rbac.authorization.k8s.io
- path: subjects
  kind: ClusterRoleBinding
  group: rbac.authorization.k8s.io
`

	expected := input

	output := runNamespaceTransformer(t, config, input)
	if output != expected {
		fmt.Println("Actual:")
		fmt.Println(output)
		fmt.Println("===")
		fmt.Println("Expected:")
		fmt.Println(expected)
		t.Fatalf("Actual doesn't equal to expected")
	}
}

func TestNamespaceTransformerCRD(t *testing.T) {
	config := `
namespace: test
fieldSpecs:
- path: metadata/namespace
  create: true
- path: data/namespace
  group: foo
  version: v1
  create: true
`

	input := `apiVersion: foo/v1
kind: bar
metadata:
  name: whatever
---
apiVersion: bar/v1
kind: foo
metadata:
  name: whatever
`

	expected := `apiVersion: foo/v1
data:
  namespace: test
kind: bar
metadata:
  name: whatever
  namespace: test
---
apiVersion: bar/v1
kind: foo
metadata:
  name: whatever
  namespace: test
`

	output := runNamespaceTransformer(t, config, input)
	if output != expected {
		fmt.Println("Actual:")
		fmt.Println(output)
		fmt.Println("===")
		fmt.Println("Expected:")
		fmt.Println(expected)
		t.Fatalf("Actual doesn't equal to expected")
	}
}

func TestNamespaceTransformerDependsOn(t *testing.T) {
	config := `
namespace: test
`
	testCases := []struct {
		TestName string
		Input    string
		Expected string
	}{
		{
			TestName: "update depends-on annotation if referenced resource is core, namespaced, and included",
			Input: `apiVersion: v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/depends-on: /namespaces/default/StatefulSet/wordpress-mysql
  labels:
    app: wordpress
  name: wordpress
  namespace: default
---
apiVersion: v1
kind: StatefulSet
metadata:
  labels:
    app: wordpress
  name: wordpress-mysql
  namespace: default
`,
			Expected: `apiVersion: v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/depends-on: /namespaces/test/StatefulSet/wordpress-mysql
  labels:
    app: wordpress
  name: wordpress
  namespace: test
---
apiVersion: v1
kind: StatefulSet
metadata:
  labels:
    app: wordpress
  name: wordpress-mysql
  namespace: test
`,
		},
		{
			TestName: "update depends-on annotation if referenced resource is namespaced and included",
			Input: `apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/depends-on: apps/namespaces/default/StatefulSet/wordpress-mysql
  labels:
    app: wordpress
  name: wordpress
  namespace: default
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    app: wordpress
  name: wordpress-mysql
  namespace: default
`,
			Expected: `apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/depends-on: apps/namespaces/test/StatefulSet/wordpress-mysql
  labels:
    app: wordpress
  name: wordpress
  namespace: test
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    app: wordpress
  name: wordpress-mysql
  namespace: test
`,
		},
		{
			TestName: "not update depends-on annotation if referenced resource is namespaced but not included",
			Input: `apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/depends-on: apps/namespaces/default/StatefulSet/wordpress-mysql
  labels:
    app: wordpress
  name: wordpress
  namespace: default
`,
			Expected: `apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/depends-on: apps/namespaces/default/StatefulSet/wordpress-mysql
  labels:
    app: wordpress
  name: wordpress
  namespace: test
`,
		},
		{
			TestName: "not update depends-on annotation referencing a cluster scoped resource",
			Input: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  annotations:
    config.kubernetes.io/depends-on: rbac.authorization.k8s.io/ClusterRole/secret-reader
  name: read-secrets-global
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: secret-reader
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: Group
  name: admin
`,
			Expected: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  annotations:
    config.kubernetes.io/depends-on: rbac.authorization.k8s.io/ClusterRole/secret-reader
  name: read-secrets-global
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: secret-reader
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: Group
  name: admin
`,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("namespace transformer should %s", tc.TestName), func(t *testing.T) {
			actual := runNamespaceTransformer(t, config, tc.Input)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}

func TestSetDependsOnNamespace(t *testing.T) {
	testCases := []struct {
		TestName       string
		Input          string
		ResourceLookup resourceLookup
		Expected       string
		IsSet          bool
	}{
		{
			TestName: "set the namespace portion for a namespaced resource which was included in the input",
			Input:    "group/namespaces/ns/kind/name",
			ResourceLookup: resourceLookup{
				resourceMap: map[resourceKey]bool{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					}: true,
				},
			},
			Expected: "group/namespaces/new-ns/kind/name",
			IsSet:    true,
		},
		{
			TestName: "set the namespace portion for a core namespaced resource which was included in the input",
			Input:    "/namespaces/ns/kind/name",
			ResourceLookup: resourceLookup{
				resourceMap: map[resourceKey]bool{
					{
						Group: "",
						Kind:  "kind",
						Name:  "name",
					}: true,
				},
			},
			Expected: "/namespaces/new-ns/kind/name",
			IsSet:    true,
		},
		{
			TestName: "not set the namespace portion for a namespaced resource which was not included in the input",
			Input:    "group/namespaces/ns/kind/name",
			Expected: "group/namespaces/ns/kind/name",
			IsSet:    false,
		},
		{
			TestName: "not set namespace for cluster scoped resource",
			Input:    "group/kind/name",
			Expected: "group/kind/name",
			IsSet:    false,
		},
		{
			TestName: "not set namespace for resource with unexpected prefix",
			Input:    "some-prefix/group/namespaces/ns/kind/name",
			Expected: "some-prefix/group/namespaces/ns/kind/name",
			IsSet:    false,
		},
		{
			TestName: "not set namespace for resource with unexpected suffix",
			Input:    "group/namespaces/ns/kind/name/some-suffix",
			Expected: "group/namespaces/ns/kind/name/some-suffix",
			IsSet:    false,
		},
		{
			TestName: "not set namespace for resource without the 'namespaces' substring",
			Input:    "group/oops/ns/kind/name/some-suffix",
			Expected: "group/oops/ns/kind/name/some-suffix",
			IsSet:    false,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("setDependsOnNamespaces should %s", tc.TestName), func(t *testing.T) {
			myPlugin := plugin{
				Namespace:           "new-ns",
				inputResourceLookup: tc.ResourceLookup,
			}
			actual, set := myPlugin.setDependsOnNamespace(tc.Input)
			assert.Equal(t, tc.IsSet, set)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}

package main

import (
	"fmt"
	"testing"
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
	if len(plugin.FieldSpecs) == 0 {
		plugin.FieldSpecs = defaultConfig
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

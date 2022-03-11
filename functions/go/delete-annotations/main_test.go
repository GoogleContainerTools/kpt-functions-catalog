package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var tests = []struct {
	name           string
	config         string
	input          string
	expectedResult []string
	errMsg         string
}{
	{
		name: "delete one doc-gen annotation",
		input: `apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sNoEnvVarSecrets
metadata:
  name: no-secrets-as-env-vars
  annotations:
  # This constraint is not certified by CIS.
    description: "Prohibits secrets as environment variables in container definitions; instead, use mounted secret files in data volumes."
    bundles.validator.forsetisecurity.org/cis-k8s-v1.5.1: 5.4.1
    policy.library/doc-gen: "do_not_document"
  spec:
    enforcementAction: dryrun
    match:
      excludedNamespaces:
      - config-management-system
      - gke-connect
`,
		config: `data:
  annotationKeys: policy.library/doc-gen
`,
		expectedResult: []string{
			"The following annotations were deleted from the resources",
			"Annonation: [policy.library/doc-gen] removed from resource: [no-secrets-as-env-vars]",
		},
	},
	{
		name: "delete multiple annotations",
		input: `apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sNoEnvVarSecrets
metadata:
  name: no-secrets-as-env-vars
  annotations:
  # This constraint is not certified by CIS.
    description: "Prohibits secrets as environment variables in container definitions; instead, use mounted secret files in data volumes."
    bundles.validator.forsetisecurity.org/cis-k8s-v1.5.1: 5.4.1
    policy.library/doc-gen: "do_not_document"
    another.annotation.to.delete: "some_value"
  spec:
    enforcementAction: dryrun 
    match:
      excludedNamespaces:
      - config-management-system
      - gke-connect
`,
		config: `data:
  annotationKeys: policy.library/doc-gen,another.annotation.to.delete
`,
		expectedResult: []string{
			"The following annotations were deleted from the resources",
			"Annonation: [policy.library/doc-gen] removed from resource: [no-secrets-as-env-vars]",
			"Annonation: [another.annotation.to.delete] removed from resource: [no-secrets-as-env-vars]",
		},
	},
	{
		name: "no annotations found",
		input: `apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sNoEnvVarSecrets
metadata:
  name: no-secrets-as-env-vars
  annotations:
  # This constraint is not certified by CIS.
    description: "Prohibits secrets as environment variables in container definitions; instead, use mounted secret files in data volumes."
    bundles.validator.forsetisecurity.org/cis-k8s-v1.5.1: 5.4.1
  spec:
    enforcementAction: dryrun 
    match:
      excludedNamespaces:
      - config-management-system
      - gke-connect
`,
		config: `data:
  annotationKeys: policy.library/doc-gen,another.annotation.to.delete
`,
		expectedResult: []string{
			"None of the resources had the provided annotations to delete",
		},
	},
}

func TestPolicyResources(t *testing.T) {
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {

			var annotationKeys string
			fcNode, err := yaml.Parse(test.config)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			err = getAnnotationKeys(fcNode, &annotationKeys)
			if err != nil {
				require.Equal(t, test.expectedResult[0], err.Error())
				return
			}

			resource, err := yaml.Parse(test.input)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			resources := append([]*yaml.RNode{}, resource)
			items, err := processResources(resources, annotationKeys)
			if err != nil {
				t.Errorf("Error when calling processResources %s", err.Error())
			}

			for j := range items {
				require.Equal(t, test.expectedResult[j], items[j].Message)
			}
		})
	}
}

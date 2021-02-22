package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func runPrefixTransformerE(config, input string) (string, error) {
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
	tc, err := getDefaultConfig()
	if err != nil {
		return "", err
	}
	plugin.FieldSpecs = append(plugin.FieldSpecs, tc.FieldSpecs...)
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

func runPrefixTransformer(t *testing.T, config, input string) string {
	s, err := runPrefixTransformerE(config, input)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestPrefixTransformer(t *testing.T) {
	config := `
prefix: dev-
`

	input := `apiVersion: v1
kind: Namespace
metadata:
  name: apple
---
apiVersion: v1
kind: Service
metadata:
  name: apple
spec:
  ports:
  - port: 7002
---
apiVersion: v1
kind: MyResource
metadata:
  name: crd
`

	expected := `apiVersion: v1
kind: Namespace
metadata:
  name: apple
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    config.kubernetes.io/prefixes: dev-
    config.kubernetes.io/previousNames: apple
    config.kubernetes.io/previousNamespaces: default
  name: dev-apple
spec:
  ports:
  - port: 7002
---
apiVersion: v1
kind: MyResource
metadata:
  annotations:
    config.kubernetes.io/prefixes: dev-
    config.kubernetes.io/previousNames: crd
    config.kubernetes.io/previousNamespaces: default
  name: dev-crd
`

	output := runPrefixTransformer(t, config, input)
	assert.EqualValues(t, expected, output, "Actual doesn't equal to expected")
}

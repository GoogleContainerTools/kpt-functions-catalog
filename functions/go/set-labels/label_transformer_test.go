package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func runLabelTransformerE(config, input string) (string, error) {
	resmapFactory := newResMapFactory()
	resMap, err := resmapFactory.NewResMapFromBytes([]byte(input))
	if err != nil {
		return "", err
	}

	var plugin *plugin = &KustomizePlugin
	plugin.Results = nil
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

func runLabelTransformer(t *testing.T, config, input string) string {
	s, err := runLabelTransformerE(config, input)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestLabelTransformer(t *testing.T) {
	config := `
labels:
  app: myApp
  quotedBoolean: "true"
  quotedFruit: "peach"
  unquotedBoolean: true
  env: production
fieldSpecs:
- path: spec/selector
  create: true
  version: v1
  kind: Service
- path: metadata/labels
  create: true
- path: spec/selector/matchLabels
  create: true
  kind: Deployment
- path: spec/template/metadata/labels
  create: true
  kind: Deployment
`

	input := `
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mungebot
  labels:
    app: mungebot
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: mungebot
    spec:
      containers:
      - name: nginx
        image: nginx
`

	expected := `apiVersion: v1
kind: Service
metadata:
  labels:
    app: myApp
    env: production
    quotedBoolean: "true"
    quotedFruit: peach
    unquotedBoolean: "true"
  name: myService
spec:
  ports:
  - port: 7002
  selector:
    app: myApp
    env: production
    quotedBoolean: "true"
    quotedFruit: peach
    unquotedBoolean: "true"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: myApp
    env: production
    quotedBoolean: "true"
    quotedFruit: peach
    unquotedBoolean: "true"
  name: mungebot
spec:
  replicas: 1
  selector:
    matchLabels:
      app: myApp
      env: production
      quotedBoolean: "true"
      quotedFruit: peach
      unquotedBoolean: "true"
  template:
    metadata:
      labels:
        app: myApp
        env: production
        quotedBoolean: "true"
        quotedFruit: peach
        unquotedBoolean: "true"
    spec:
      containers:
      - image: nginx
        name: nginx
`

	output := runLabelTransformer(t, config, input)
	if output != expected {
		fmt.Println("Actual:")
		fmt.Println(output)
		fmt.Println("===")
		fmt.Println("Expected:")
		fmt.Println(expected)
		t.Fatalf("Actual doesn't equal to expected")
	}
}

func TestLabelTransformerIdempotence(t *testing.T) {
	config := `
labels:
  foo: bar
fieldSpecs:
- kind: ConfigMap
  path: metadata/labels
  create: true
`
	input := `
apiVersion: apps/v1
kind: ConfigMap
metadata:
  name: whatever
data: {}
`

	expected := `apiVersion: apps/v1
data: {}
kind: ConfigMap
metadata:
  labels:
    foo: bar
  name: whatever
`

	// do the transformation twice
	output := runLabelTransformer(t, config, input)
	output = runLabelTransformer(t, config, output)
	if output != expected {
		fmt.Println("Actual:")
		fmt.Println(output)
		fmt.Println("===")
		fmt.Println("Expected:")
		fmt.Println(expected)
		t.Fatalf("Actual doesn't equal to expected")
	}
}

func TestAnnotationsTransformerResults(t *testing.T) {
	config := `
labels:
  foo: bar
  baz: bat
fieldSpecs:
- kind: ConfigMap
  path: metadata/labels
  create: true
`
	input := `
apiVersion: apps/v1
kind: ConfigMap
metadata:
  annotations:
    internal.config.kubernetes.io/path: foo.yaml
    internal.config.kubernetes.io/index: 0
  name: vilgefortz
data: {}
---
apiVersion: apps/v1
kind: ConfigMap
metadata:
  annotations:
    internal.config.kubernetes.io/path: bar.yaml
    internal.config.kubernetes.io/index: 1
  name: triss
data: {}
`
	expectedResults := LabelResults{
		{
			FilePath:  "foo.yaml",
			FileIndex: "0",
			FieldPath: "metadata.labels",
		}: {"foo": "bar", "baz": "bat"},
		{
			FilePath:  "bar.yaml",
			FileIndex: "1",
			FieldPath: "metadata.labels",
		}: {"foo": "bar", "baz": "bat"},
	}
	runLabelTransformer(t, config, input)
	assert.Equal(t, expectedResults, KustomizePlugin.Results)
}

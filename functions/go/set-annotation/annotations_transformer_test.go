// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"testing"
)

func runAnnotationTransformerE(config, input string) (string, error) {
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

func runAnnotationTransformer(t *testing.T, config, input string) string {
	s, err := runAnnotationTransformerE(config, input)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestAnnotationsTransformer(t *testing.T) {
	config := `
annotations:
  app: myApp
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
  annotations:
    app: myApp
  name: myService
spec:
  ports:
  - port: 7002
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    app: myApp
  labels:
    app: mungebot
  name: mungebot
spec:
  replicas: 1
  template:
    metadata:
      annotations:
        app: myApp
      labels:
        app: mungebot
    spec:
      containers:
      - image: nginx
        name: nginx
`
	output := runAnnotationTransformer(t, config, input)
	if output != expected {
		fmt.Println("Actual:")
		fmt.Println(output)
		fmt.Println("===")
		fmt.Println("Expected:")
		fmt.Println(expected)
		t.Fatalf("Actual doesn't equal to expected")
	}
}

func TestAnnotationsTransformerIdempotence(t *testing.T) {
	config := `
annotations:
  app: myApp
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
  annotations:
    app: myApp
  name: myService
spec:
  ports:
  - port: 7002
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    app: myApp
  labels:
    app: mungebot
  name: mungebot
spec:
  replicas: 1
  template:
    metadata:
      annotations:
        app: myApp
      labels:
        app: mungebot
    spec:
      containers:
      - image: nginx
        name: nginx
`
	// do the transformation twice
	output := runAnnotationTransformer(t, config, input)
	output = runAnnotationTransformer(t, config, output)
	if output != expected {
		fmt.Println("Actual:")
		fmt.Println(output)
		fmt.Println("===")
		fmt.Println("Expected:")
		fmt.Println(expected)
		t.Fatalf("Actual doesn't equal to expected")
	}
}

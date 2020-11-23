package main_test

import (
	"bytes"
	"os/exec"
	"testing"
)

func TestFunction(t *testing.T) {
	input := `
apiVersion: config.kubernetes.io/v1beta1
kind: ResourceList
functionConfig:
  apiVersion: foo-corp.com/v1
  kind: FulfillmentCenter
  metadata:
    name: staging
    annotations:
      config.kubernetes.io/function: |
        container:
          image: gcr.io/example/foo:v1.0.0
  data:
    namespace: foo
    fieldSpecs:
    - path: metadata/namespace
      create: true
items:
  - apiVersion: apps/v1
    kind: foobar
    metadata:
      name: whatever
`
	expected := `apiVersion: config.kubernetes.io/v1beta1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: foobar
  metadata:
    name: whatever
    namespace: foo
functionConfig:
  apiVersion: foo-corp.com/v1
  kind: FulfillmentCenter
  metadata:
    name: staging
    annotations:
      config.kubernetes.io/function: |
        container:
          image: gcr.io/example/foo:v1.0.0
  data:
    namespace: foo
    fieldSpecs:
    - path: metadata/namespace
      create: true
`
	cmd := exec.Command("go", "run", ".")
	cmd.Stdin = bytes.NewBufferString(input)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Exec failed: %s", out)
	}

	output := string(out)

	if output != expected {
		t.Fatalf("Output doesn't match expected.\nOutput:\n%s\nExpected:\n%s", output, expected)
	}
}

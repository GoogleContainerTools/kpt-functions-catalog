package transformer

import (
	"fmt"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"sigs.k8s.io/yaml"
	"testing"
)

func runTest(t *testing.T, config, input, expected string) {
	var sl SetLabels
	err := yaml.Unmarshal([]byte(config), &sl)
	if err != nil {
		return
	}
	in, _ := fn.ParseKubeObject([]byte(input))
	sl.Run(nil, nil, fn.KubeObjects{in})
	//fmt.Println(ctx)
	exp, _ := fn.ParseKubeObject([]byte(expected))
	if exp.String() != in.String() {
		fmt.Println("Actual:")
		fmt.Println(in)
		fmt.Println("===")
		fmt.Println("Expected:")
		fmt.Println(exp)
		t.Fatalf("Actual doesn't equal to expected")
	}
}

func Test_simple(t *testing.T) {
	config := `
labels:
  color: orange
  name: apple
`
	input := `
apiVersion: apps/v1
kind: ConfigMap
metadata:
  name: whatever
`
	expected := `
apiVersion: apps/v1
kind: ConfigMap
metadata:
  name: whatever
  labels:
    color: orange
    name: apple
`
	runTest(t, config, input, expected)
}

func Test_Additional_Field(t *testing.T) {
	config := `
labels:
  color: orange
  name: apple
additionalLabelFields:
  - path: data/selector
    kind: ConfigMap
    create: true
`
	input := `
apiVersion: apps/v1
kind: ConfigMap
metadata:
  name: whatever
`
	expected := `
apiVersion: apps/v1
kind: ConfigMap
metadata:
  name: whatever
  labels:
    color: orange
    name: apple
data:
  selector:
    color: orange
    name: apple
`
	runTest(t, config, input, expected)
}

package transformer

import (
	"fmt"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"testing"
)

func TestLabelTransformer_simple_ConfigMap(t *testing.T) {
	functionConfig := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  app: myApp
  quotedBoolean: "true"
  quotedFruit: "peach"
  unquotedBoolean: true
  env: production
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
    app: myApp
    env: production
    quotedBoolean: "true"
    quotedFruit: peach
    unquotedBoolean: "true"
`

	transformer := LabelTransformer{}
	config, _ := fn.ParseKubeObject([]byte(functionConfig))
	_ = transformer.Config(config)
	result, _ := fn.ParseKubeObject([]byte(input))
	err := transformer.Transform(fn.KubeObjects{result})
	if err != nil {
		return
	}
	fmt.Println(transformer.Results)
	exp, _ := fn.ParseKubeObject([]byte(expected))

	if exp.String() != result.String() {
		fmt.Println("Actual:")
		fmt.Println(result)
		fmt.Println("===")
		fmt.Println("Expected:")
		fmt.Println(exp)
		t.Fatalf("Actual doesn't equal to expected")
	}
}

func TestLabelTransformer_simple_ConfigMap2(t *testing.T) {
	functionConfig := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  app: myApp
  quotedBoolean: "true"
  quotedFruit: "peach"
  unquotedBoolean: true
  env: production
`
	input := `
apiVersion: apps/v1
kind: ConfigMap
metadata:
  name: whatever
  labels:
    extra: nil
    env: dev
`

	expected := `apiVersion: apps/v1
kind: ConfigMap
metadata:
  name: whatever
  labels:
    extra: nil
    env: production
    app: myApp
    quotedBoolean: "true"
    quotedFruit: peach
    unquotedBoolean: "true"`

	transformer := LabelTransformer{}
	config, _ := fn.ParseKubeObject([]byte(functionConfig))
	_ = transformer.Config(config)
	result, _ := fn.ParseKubeObject([]byte(input))
	err := transformer.Transform(fn.KubeObjects{result})
	if err != nil {
		return
	}
	exp, _ := fn.ParseKubeObject([]byte(expected))

	if exp.String() != result.String() {
		fmt.Println("Actual:")
		fmt.Println(result)
		fmt.Println("===")
		fmt.Println("Expected:")
		fmt.Println(exp)
		t.Fatalf("Actual doesn't equal to expected")
	}
}

func TestLabelTransformer_simple_ConfigFile(t *testing.T) {
	functionConfig := `
apiVersion: fn.kpt.dev/v1alpha1
kind: SetLabels
metadata:
  name: my-config
labels:
  color: orange
  fruit: apple
`

	input := `
apiVersion: apps/v1
kind: MyResource
metadata:
  name: whatever
  labels:
    extra: nil
    env: dev
spec:
  selector:
    labels:
      fruit: apple
      name: jemma
`

	expected := `
apiVersion: apps/v1
kind: MyResource
metadata:
  name: whatever
  labels:
    extra: nil
    env: dev
    color: orange
    fruit: apple
spec:
  selector:
    labels:
      fruit: apple
      name: jemma
`

	transformer := LabelTransformer{}
	config, _ := fn.ParseKubeObject([]byte(functionConfig))
	_ = transformer.Config(config)
	result, _ := fn.ParseKubeObject([]byte(input))
	err := transformer.Transform(fn.KubeObjects{result})
	if err != nil {
		return
	}
	//fmt.Println(result.String())
	exp, _ := fn.ParseKubeObject([]byte(expected))
	if exp.String() != result.String() {
		fmt.Println("Actual:")
		fmt.Println(result)
		fmt.Println("===")
		fmt.Println("Expected:")
		fmt.Println(exp)
		t.Fatalf("Actual doesn't equal to expected")
	}
}

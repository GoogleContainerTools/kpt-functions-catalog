package transformer

import (
	"fmt"
	"testing"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

func TestLabelTransformer_ConfigMap_Service(t *testing.T) {
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
apiVersion: v1
kind: Service
metadata:
  name: whatever
spec:
  selector:
    a: b
`

	expected := `
apiVersion: v1
kind: Service
metadata:
  name: whatever
  labels:
    app: myApp
    env: production
    quotedBoolean: "true"
    quotedFruit: peach
    unquotedBoolean: "true"
spec:
  selector:
    a: b
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

func TestLabelTransformer_ConfigMap_Slice(t *testing.T) {
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
apiVersion: apps/
kind: StatefulSet
metadata:
  name: my-config
spec:
  volumeClaimTemplates:
    - metadata:
        labels:
          testkey: testvalue
`

	expected := `
apiVersion: apps/
kind: StatefulSet
metadata:
  name: my-config
  labels:
    app: myApp
    env: production
    quotedBoolean: "true"
    quotedFruit: peach
    unquotedBoolean: "true"
spec:
  volumeClaimTemplates:
  - metadata:
      labels:
        testkey: testvalue
        app: myApp
        env: production
        quotedBoolean: "true"
        quotedFruit: peach
        unquotedBoolean: "true"
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
`

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

func TestLabelTransformer_simple_ConfigMap_Result(t *testing.T) {
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
    env: production
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

	expectedResult := "set labels: {\"app\":\"myApp\",\"quotedBoolean\":\"true\",\"quotedFruit\":\"peach\",\"unquotedBoolean\":\"true\"}"
	if transformer.Results[0].Message != expectedResult {
		t.Fatalf("Actual doesn't equal to expected")
	}

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

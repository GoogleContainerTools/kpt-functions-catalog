package transformer

import (
	"fmt"
	"testing"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"github.com/google/go-cmp/cmp"
)

func generateResourceList(functionConfig string, items []string) *fn.ResourceList {
	// generate recourse list, config function config, then upsert items
	rl := &fn.ResourceList{}
	config, _ := fn.ParseKubeObject([]byte(functionConfig))
	rl.FunctionConfig = config
	for _, item := range items {
		itemObj, _ := fn.ParseKubeObject([]byte(item))
		if err := rl.UpsertObjectToItems(itemObj, nil, false); err != nil {
			panic("add items failed")
		}
	}
	return rl
}

func runTest(functionConfig string, items []string, expectedItems []string, expMsg []string) bool {
	rl := generateResourceList(functionConfig, items)
	_, err := SetLabels(rl)
	if err != nil {
		return false
	}
	// compare items
	if expectedItems != nil {
		for idx, item := range expectedItems {
			if !compareString(rl.Items[idx].String(), item) {
				return false
			}
		}
	}
	if expMsg != nil {
		msgIdx := 0
		for idx := 0; idx < rl.Items.Len(); idx++ {
			if !compareString(rl.Results[idx].Message, expMsg[msgIdx]) {
				return false
			}
			msgIdx++
		}
	}

	return true
}

func compareString(actual string, expected string) bool {
	if !cmp.Equal(actual, expected) {
		fmt.Println("Actual:")
		fmt.Println(actual)
		fmt.Println("===")
		fmt.Println("Expected:")
		fmt.Println(expected)
		fmt.Println(cmp.Diff(actual, expected))
		return false
	}
	return true
}

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

	expected := `apiVersion: v1
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

	if !runTest(functionConfig, []string{input}, []string{expected}, nil) {
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

	expected := `apiVersion: apps/
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

	if !runTest(functionConfig, []string{input}, []string{expected}, nil) {
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
    unquotedBoolean: "true"
`
	expectedLogResult := `set labels: {"app":"myApp","env":"production","quotedBoolean":"true","quotedFruit":"peach","unquotedBoolean":"true"}`

	if !runTest(functionConfig, []string{input}, []string{expected}, []string{expectedLogResult}) {
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
    unquotedBoolean: "true"
`

	expectedLogResult := `set labels: {"app":"myApp","quotedBoolean":"true","quotedFruit":"peach","unquotedBoolean":"true"}`

	if !runTest(functionConfig, []string{input}, []string{expected}, []string{expectedLogResult}) {
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

	expected := `apiVersion: apps/v1
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

	if !runTest(functionConfig, []string{input}, []string{expected}, nil) {
		t.Fatalf("Actual doesn't equal to expected")
	}
}

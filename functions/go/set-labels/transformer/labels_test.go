package transformer

import (
	"testing"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"github.com/stretchr/testify/assert"
)

func TestSetLabels(t *testing.T) {
	configMap := `
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

	setLabelsConfig := `
apiVersion: fn.kpt.dev/v1alpha1
kind: SetLabels
metadata:
  name: my-config
labels:
  app: myApp
  quotedBoolean: "true"
  unquotedBoolean: true
  env: production
  quotedFruit: "peach"
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

	sliceInput := `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: my-config
spec:
  volumeClaimTemplates:
    - metadata:
        labels:
          testkey: testvalue
          app: testvalue1
          env: testvalue2
          quotedBoolean: testvalue3
          quotedFruit: testvalue4
          unquotedBoolean: testvalue5
`

	sameLabelInput := `
apiVersion: apps/v1
kind: ConfigMap
metadata:
  name: whatever
  labels:
    extra: nil
    env: production
`

	sameLabelExpected := `apiVersion: apps/v1
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

	sameLableLogResult := []string{"set labels {app : myApp} for 1 times", "set labels {quotedBoolean : true} for 1 times", "set labels {quotedFruit : peach} for 1 times", "set labels {unquotedBoolean : true} for 1 times"}

	logResult := []string{"set labels {app : myApp} for 2 times", "set labels {env : production} for 2 times", "set labels {quotedBoolean : true} for 2 times", "set labels {quotedFruit : peach} for 2 times", "set labels {unquotedBoolean : true} for 2 times"}

	sliceExpected := `apiVersion: apps/v1
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

	var testCases = map[string]struct {
		resourcelist *fn.ResourceList
		expected     []*fn.KubeObject
		logResult    []string
	}{
		"Update resources with ConfigMap": {
			resourcelist: generateResourceList(configMap, []string{input}),
			expected:     generateExpectedResult([]string{expected}),
			logResult:    logResult,
		},
		"Update resources that contains slice structure with configMap, ": {
			resourcelist: generateResourceList(configMap, []string{sliceInput}),
			expected:     generateExpectedResult([]string{sliceExpected}),
		},
		"Update resources using setLabel kind": {
			resourcelist: generateResourceList(setLabelsConfig, []string{input}),
			expected:     generateExpectedResult([]string{expected}),
		},
		"Resource has the same label as configMap, log results omit this log": {
			resourcelist: generateResourceList(configMap, []string{sameLabelInput}),
			expected:     generateExpectedResult([]string{sameLabelExpected}),
			logResult:    sameLableLogResult,
		},
	}

	for testName, data := range testCases {
		success, _ := SetLabels(data.resourcelist)
		if success != true {
			t.Fatalf("Set labels error")
		}

		for idx, item := range data.expected {
			assert.Equal(t, item.String(), data.resourcelist.Items[idx].String(), testName)
		}

		if data.logResult != nil {
			for idx, item := range data.logResult {
				assert.Equal(t, item, data.resourcelist.Results[idx].Message, testName+" log error")
			}
		}
	}
}

// generateExpectedResult parse the expected from string to kubeObject
func generateExpectedResult(expected []string) []*fn.KubeObject {
	var res []*fn.KubeObject
	for _, exp := range expected {
		obj, _ := fn.ParseKubeObject([]byte(exp))
		res = append(res, obj)
	}
	return res
}

// generateResourceList generate recourse list, config function config, then upsert items
func generateResourceList(functionConfig string, items []string) *fn.ResourceList {
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

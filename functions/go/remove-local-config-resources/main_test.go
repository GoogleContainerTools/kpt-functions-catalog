package main

import (
	"fmt"
	"testing"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	// "sigs.k8s.io/kustomize/kyaml/fn/framework"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const testDir = "testdata"

type PrunerTest struct {
	Name           string
	ResFiles       []string
	ExpectedResult string
}

var Tests = []PrunerTest{
	{
		Name:           "Multiple Pruned Resources",
		ResFiles:       []string{"applied.yaml", "local-01.yaml", "local-02.yaml"},
		ExpectedResult: "Resources Pruned: [Count: 2, Names: {sample-hierarchy-01, sample-hierarchy-02}]",
	},
	{
		Name:           "Single Pruned Resource",
		ResFiles:       []string{"applied.yaml", "local-01.yaml"},
		ExpectedResult: "Resources Pruned: [Count: 1, Names: {sample-hierarchy-01}]",
	},
	{
		Name:           "No Pruned Resource",
		ResFiles:       []string{"applied.yaml"},
		ExpectedResult: "Resources Pruned: [Count: 0, Names: {local resources not found}]",
	},
}

func TestPrunedResources(t *testing.T) {

	for i := range Tests {
		test := Tests[i]
		t.Run(test.Name, func(t *testing.T) {
			rl := &framework.ResourceList{}

			err := loadYAMLs(rl, test.ResFiles...)
			if err != nil {
				t.Errorf("Error when loading yaml files %s", err.Error())
				return
			}

			var items []framework.ResultItem

			items, err = ProcessResources(rl)
			if err != nil {
				t.Errorf("Error when calling ProcessResources %s", err.Error())
			}

			println(items[0].Message)

			resultMessage := items[0].Message

			if !assert.Equal(t, resultMessage, test.ExpectedResult) {
				t.FailNow()
			}
		})
	}
}

func loadYAMLs(rl *framework.ResourceList, filenames ...string) error {

	for _, filename := range filenames {
		node, err := yaml.ReadFile(fmt.Sprintf("%s/%s", testDir, filename))
		if err != nil {
			return err
		}
		rl.Items = append(rl.Items, node)
	}

	return nil
}

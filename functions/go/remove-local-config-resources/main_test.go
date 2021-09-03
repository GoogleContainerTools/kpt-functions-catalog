package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const testDir = "testdata"

type PrunerTest struct {
	Name           string
	ResFiles       []string
	ExpectedResult []string
}

var Tests = []PrunerTest{
	{
		Name:     "Multiple Pruned Resources",
		ResFiles: []string{"applied.yaml", "local-01.yaml", "local-02.yaml"},
		ExpectedResult: []string{
			"Number of resources pruned: 2",
			"Resource name: [sample-hierarchy-01]",
			"Resource name: [sample-hierarchy-02]",
		},
	},
	{
		Name:     "Single Pruned Resource",
		ResFiles: []string{"applied.yaml", "local-01.yaml"},
		ExpectedResult: []string{
			"Number of resources pruned: 1",
			"Resource name: [sample-hierarchy-01]",
		},
	},
	{
		Name:     "No Pruned Resource",
		ResFiles: []string{"applied.yaml"},
		ExpectedResult: []string{
			"Found no resources to prune with the local config annotation",
		},
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

			items, err = processResources(rl)
			if err != nil {
				t.Errorf("Error when calling processResources %s", err.Error())
			}

			for j := range items {
				require.Equal(t, test.ExpectedResult[j], items[j].Message)
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

// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package terraformgenerator

import (
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/fn/sdk"
	"sigs.k8s.io/kustomize/kyaml/fn/sdk/testutil"
)

const testDir = "testdata"

type TerraformTest struct {
	Name string
}

var testCases = []TerraformTest{
	{
		Name: "team",
	},
	{
		Name: "empty",
	},
	{
		Name: "other_resources",
	},
}

func TestTerraformGeneration(t *testing.T) {
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			require := require.New(t)
			inDir := path.Join("..", testDir, tt.Name, "input")
			outDir := path.Join("..", testDir, tt.Name, "output")

			actualRL, err := testutil.ResourceListFromDirectory(inDir, "")
			require.NoError(err)

			expectedRL, err := testutil.ResourceListFromDirectory(outDir, "")
			require.NoError(err)

			// append input items to output
			expectedRL.Items = append(expectedRL.Items, actualRL.Items...)

			err = Processor(actualRL)
			require.NoError(err)

			actualTerraform, err := findTerraform(actualRL)
			require.NoError(err)
			expectedTerraform, err := findTerraform(expectedRL)
			require.NoError(err)
			require.EqualValues(fmt.Sprint(expectedTerraform), fmt.Sprint(actualTerraform))

			// final check on yaml
			actualYAML, err := actualRL.ToYAML()
			require.NoError(err)
			require.NotEmpty(actualYAML)
			expectedYAML, err := expectedRL.ToYAML()
			require.NoError(err)
			require.NotEmpty(expectedYAML)

			require.YAMLEqf(string(expectedYAML), string(actualYAML), "output yaml doesn't match")
		})
	}
}

func findTerraform(rl *sdk.ResourceList) (map[string]string, error) {
	for _, obj := range rl.Items {
		if obj.Name() == "terraform" {
			values := make(map[string]string)
			_, err := obj.Get(&values, "data")
			if err != nil {
				return nil, err
			}
			return values, nil
		}
	}
	return nil, fmt.Errorf("No terraform file found.")
}

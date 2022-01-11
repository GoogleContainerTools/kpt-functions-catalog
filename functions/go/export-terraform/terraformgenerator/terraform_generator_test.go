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
	"io/ioutil"
	"path"
	"testing"

	sdk "github.com/GoogleContainerTools/kpt-functions-catalog/thirdparty/kyaml/fnsdk"
	"github.com/GoogleContainerTools/kpt-functions-catalog/thirdparty/kyaml/fnsdk/testutil"
	"github.com/stretchr/testify/require"
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

			// We convert the output ResourceList to individual resource files in the
			// file system first.
			// In the next step, ideally, we should compare the content in the
			// expected directory and the actual directory. But I haven't found a good
			// golang pkg for that yet. Maybe we can rely on some external tool (e.g.
			// diff command) to do it. This will be addressed in the next iteration.
			// The workaround is that we read the resource files as a ResourceList and
			// then compare this ResourceList with the expected ResourceList.
			tmpDir, err := ioutil.TempDir("", "export-terraform-test-*")
			fmt.Println(tmpDir)
			//defer os.RemoveAll(tmpDir)
			require.NoError(err)
			err = testutil.ResourceListToDirectory(actualRL, tmpDir)
			require.NoError(err)

			tmpDirRL, err := testutil.ResourceListFromDirectory(tmpDir, "")
			require.NoError(err)

			// final check on yaml
			tmpDirYAML, err := tmpDirRL.ToYAML()
			require.NoError(err)
			require.NotEmpty(tmpDirYAML)
			expectedYAML, err := expectedRL.ToYAML()
			require.NoError(err)
			require.NotEmpty(expectedYAML)

			require.YAMLEqf(string(expectedYAML), string(tmpDirYAML), "output yaml doesn't match")
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

package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/contrib/functions/go/blueprint-docs/docs"
	"github.com/GoogleContainerTools/kpt-functions-catalog/contrib/functions/go/blueprint-docs/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const defaultReadmePath = "/tmp/README.md"
const defaultRepoPath = "https://github.com/GoogleCloudPlatform/blueprints.git/catalog/"

//nolint
func main() {
	rp := ReadmeProcessor{}
	cmd := command.Build(&rp, command.StandaloneEnabled, false)

	cmd.Short = generated.GenerateBlueprintDocsShort
	cmd.Long = generated.GenerateBlueprintDocsLong
	cmd.Example = generated.GenerateBlueprintDocsExamples
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type ReadmeProcessor struct{}

func (rp *ReadmeProcessor) Process(resourceList *framework.ResourceList) error {
	resourceList.Result = &framework.Result{
		Name: "generate-blueprint-docs",
	}
	readmePath, repoPath := parseFnCfg(resourceList.FunctionConfig)
	doc, err := docs.GenerateBlueprintReadme(resourceList.Items, repoPath)
	if err != nil {
		resourceList.Result.Items = getErrorItem(err.Error())
		return err
	}
	err = ioutil.WriteFile(readmePath, []byte(doc), os.ModePerm)
	if err != nil {
		resourceList.Result.Items = getErrorItem(err.Error())
		return err
	}
	return nil
}

func parseFnCfg(r *yaml.RNode) (string, string) {
	cm := r.GetDataMap()
	readme, exists := cm["readme-path"]
	if !exists {
		readme = defaultReadmePath
	}
	repoPath, exists := cm["repo-path"]
	if !exists {
		repoPath = defaultRepoPath
	}
	return readme, repoPath

}

// getErrorItem returns the item for input error message
func getErrorItem(errMsg string) []framework.ResultItem {
	return []framework.ResultItem{
		{
			Message:  fmt.Sprintf("failed to generate doc: %s", errMsg),
			Severity: framework.Error,
		},
	}
}

package main

import (
	"errors"
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
		os.Exit(1)
	}
}

type ReadmeProcessor struct{}

func (rp *ReadmeProcessor) Process(resourceList *framework.ResourceList) error {
	resourceList.Result = &framework.Result{
		Name: "generate-blueprint-docs",
	}
	readmePath, repoPath := parseFnCfg(resourceList.FunctionConfig)

	currentDoc, err := ioutil.ReadFile(readmePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			resourceList.Result.Items = getResultItem(fmt.Sprintf("Skipping readme generation: %s", err), framework.Warning)
			return nil
		} else {
			resourceList.Result.Items = getResultItem(err.Error(), framework.Error)
		}
	}
	err = generateReadme(repoPath, readmePath, string(currentDoc), resourceList)
	if err != nil {
		resourceList.Result.Items = getResultItem(err.Error(), framework.Error)
		return err
	}
	return nil
}

func generateReadme(repoPath, readmePath, currentDoc string, resourceList *framework.ResourceList) error {
	title, generatedDoc, err := docs.GenerateBlueprintReadme(resourceList.Items, repoPath)
	if err != nil {
		return err
	}
	readme, err := docs.InsertIntoReadme(title, currentDoc, generatedDoc)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(readmePath, []byte(readme), os.ModePerm)
	if err != nil {
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

// getResultItem returns the item for input error message
func getResultItem(msg string, severity framework.Severity) []framework.ResultItem {
	return []framework.ResultItem{
		{
			Message:  fmt.Sprintf("failed to generate doc: %s", msg),
			Severity: severity,
		},
	}
}

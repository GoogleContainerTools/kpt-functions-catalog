package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-gcp-resource-ids/pkg/kpt"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func main() {
	// TODO: fn.AsMain should support an "easy mode" where it runs against a directory
	if err := fn.AsMain(fn.ResourceListProcessorFunc(Run)); err != nil {
		os.Exit(1)
	}
}

type SetGCPProject struct {
	ProjectID string `json:"projectID,omitempty"`
}

func Run(rl *fn.ResourceList) (bool, error) {
	f := SetGCPProject{}

	err := f.LoadConfig(rl.FunctionConfig)
	if err != nil {
		rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(fmt.Errorf("functionConfig error: %w", err), rl.FunctionConfig))
		return true, nil
	}

	if err := f.Transform(rl.Items); err != nil {
		rl.Results = append(rl.Results, fn.ErrorResult(err))
	}
	return true, nil
}

func (p *SetGCPProject) LoadConfig(fnConfig *fn.KubeObject) error {
	if fnConfig != nil {
		switch { //TODO: o.GroupVersionKind()
		case fnConfig.IsGVK("", "v1", "ConfigMap"):
			data := fnConfig.UpsertMap("data") // TODO: Why does GetMap fail?
			p.ProjectID = data.GetString("projectID")

		default:
			gvk := schema.GroupVersionKind{}                         // TODO: o.GroupVersionKind()
			return fmt.Errorf("unknown functionConfig Kind %v", gvk) //o.GroupVersionKind())
		}
	}

	return nil
}

func (p *SetGCPProject) GenerateProjectID(objects fn.KubeObjects) (string, error) {
	packageContext, err := kpt.FindPackageContext(objects)
	if err != nil {
		return "", err
	}

	projects := objects.Where(fn.IsGVK("resourcemanager.cnrm.cloud.google.com", "v1beta1", "Project"))
	if len(projects) == 0 {
		return "", fmt.Errorf("did not find any Project objects in package, cannot generate project id")
	}
	if len(projects) != 1 {
		// TODO: We could probably support this...
		return "", fmt.Errorf("found multiple Project objects in package, cannot generate project id")
	}

	project := projects[0]
	projectID, err := GenerateProjectID(project.GetName(), packageContext.ParentPath)
	if err != nil {
		return "", err
	}
	return projectID, nil
}

func (p *SetGCPProject) Transform(objects fn.KubeObjects) error {
	projectID := p.ProjectID
	if projectID == "" {
		// TODO: Only if we need a project id (though there aren't many cases where we don't)
		p, err := p.GenerateProjectID(objects)
		if err != nil {
			return err
		}
		projectID = p
	}

	packageContext, err := kpt.FindPackageContext(objects)
	if err != nil {
		return err
	}

	for _, object := range objects {
		if object.IsLocalConfig() {
			continue
		}
		if kpt.IsResourceGroup(object) {
			continue // Should ResourceGroup be marked as local config?
		}

		name := object.GetName()

		if object.IsGVK("resourcemanager.cnrm.cloud.google.com", "v1beta1", "Folder") {
			displayName := name
			object.SetNestedString(displayName, "spec", "displayName")
			// resourceID should be left unset to create a new resource
		}

		if object.IsGVK("resourcemanager.cnrm.cloud.google.com", "v1beta1", "Project") {
			// https://cloud.google.com/resource-manager/docs/creating-managing-projects
			// A project name can contain only letters, numbers, single quotes, hyphens, spaces,
			// or exclamation points, and must be between 4 and 30 characters.

			displayName := name
			if packageContext.ParentPath != "" {
				// TODO: Move to helper
				parentPathTokens := strings.Split(packageContext.ParentPath, "/")

				parentPathTokens = reverse(parentPathTokens)

				displayName += "-" + strings.Join(parentPathTokens, "-")
			}
			displayName = strings.ReplaceAll(displayName, ".", "-")

			if len(displayName) > 30 {
				displayName = displayName[:30]
			}

			object.SetNestedString(displayName, "spec", "name")     // name is the display name
			object.SetNestedString(projectID, "spec", "resourceID") // resourceID is the project ID (must be unique)
		}

		if object.GetAnnotation("cnrm.cloud.google.com/project-id") != "" {
			object.SetAnnotation("cnrm.cloud.google.com/project-id", projectID)
		}

		if object.IsGVK("core.cnrm.cloud.google.com", "v1beta1", "ConfigConnectorContext") {
			// TODO: ConfigConnectorContext should accept a serviceAccountRef
			googleServiceAccount, _, _ := object.NestedString("spec", "googleServiceAccount")
			if googleServiceAccount != "" {
				tokens := strings.Split(googleServiceAccount, "@")
				if len(tokens) != 2 {
					return fmt.Errorf("error parsing spec.googleServiceAccount=%q", googleServiceAccount)
				}
				if strings.HasSuffix(tokens[1], ".iam.gserviceaccount.com") {
					tokens[1] = projectID + ".iam.gserviceaccount.com"
				} else {
					return fmt.Errorf("unexpected value for spec.googleServiceAccount=%q (expected .iam.gserviceaccount.com suffix)", googleServiceAccount)
				}
				googleServiceAccount = strings.Join(tokens, "@")
				object.SetNestedString(googleServiceAccount, "spec", "googleServiceAccount")
			}
		}

		// ContainerNodePool has something sort of similar ... the resourceID should be the name without the prefix
		// This is better enforced via a "should" rule, I think
	}

	return nil
}

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-gcp-resource-ids/pkg/kpt"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	Folder = schema.GroupVersionKind{Group: "resourcemanager.cnrm.cloud.google.com", Version: "v1beta1", Kind: "Folder"}
	GCPProject = schema.GroupVersionKind{Group: "resourcemanager.cnrm.cloud.google.com", Version: "v1beta1", Kind: "Project"}
	ConfigControllerContext = schema.GroupVersionKind{Group: "core.cnrm.cloud.google.com", Version: "v1beta1", Kind: "ConfigConnectorContext"}
	)

func main() {
	// TODO: fn.AsMain should support an "easy mode" where it runs against a directory
	if err := fn.AsMain(&SetGCPProject{}); err != nil {
		os.Exit(1)
	}
}

var _ fn.Runner = &SetGCPProject{}

type SetGCPProject struct {
	ProjectID string `json:"projectID,omitempty"`
}

func (p *SetGCPProject) GenerateProjectID(objects fn.KubeObjects) (string, error) {
	packageContext, err := kpt.FindPackageContext(objects)
	if err != nil {
		return "", err
	}

	projects := objects.Where(fn.IsGroupVersionKind(GCPProject))
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


func (p *SetGCPProject) Run(ctx *fn.Context, functionConfig *fn.KubeObject, objects fn.KubeObjects) {
	projectID := p.ProjectID
	if projectID == "" {
		// TODO: Only if we need a project id (though there aren't many cases where we don't)
		var err error
		projectID, err = p.GenerateProjectID(objects)
		if err != nil {
			ctx.ResultErrAndDie(err.Error(), nil)
		}
	}

	packageContext, err := kpt.FindPackageContext(objects)
	if err != nil {
		ctx.ResultErrAndDie(err.Error(), nil)
	}
	for _, object := range objects {
		if object.IsLocalConfig() {
			continue
		}
		if kpt.IsResourceGroup(object) {
			continue // Should ResourceGroup be marked as local config?
		}
		name := object.GetName()

		if object.IsGroupVersionKind(Folder) {
			displayName := name
			object.SetNestedString(displayName, "spec", "displayName")
			// resourceID should be left unset to create a new resource
		}

		if object.IsGroupVersionKind(GCPProject) {
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


		if object.IsGroupVersionKind(ConfigControllerContext) {
			// TODO: ConfigConnectorContext should accept a serviceAccountRef
			googleServiceAccount, _, _ := object.NestedString("spec", "googleServiceAccount")
			if googleServiceAccount != "" {
				tokens := strings.Split(googleServiceAccount, "@")
				if len(tokens) != 2 {
					ctx.ResultErrAndDie(fmt.Sprintf("error parsing spec.googleServiceAccount=%q", googleServiceAccount), nil)
				}
				if strings.HasSuffix(tokens[1], ".iam.gserviceaccount.com") {
					tokens[1] = projectID + ".iam.gserviceaccount.com"
				} else {
					ctx.ResultErrAndDie(fmt.Sprintf("unexpected value for spec.googleServiceAccount=%q (expected .iam.gserviceaccount.com suffix)", googleServiceAccount), nil)
				}
				googleServiceAccount = strings.Join(tokens, "@")
				object.SetNestedString(googleServiceAccount, "spec", "googleServiceAccount")
			}
		}

		// ContainerNodePool has something sort of similar ... the resourceID should be the name without the prefix
		// This is better enforced via a "should" rule, I think
	}
}

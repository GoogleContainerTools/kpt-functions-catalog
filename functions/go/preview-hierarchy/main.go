package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	folderKind          = "Folder"
	kccAPIVersion       = "resourcemanager.cnrm.cloud.google.com/v1beta1"
	folderRefAnnotation = "cnrm.cloud.google.com/folder-ref"
	orgRefAnnotation    = "cnrm.cloud.google.com/organization-id"
	outputFlagName      = "output"
	renderFlagName      = "renderer"
)

type NodeType int64

const (
	Org NodeType = iota
	Folder
)

func (n NodeType) String() string {
	switch n {
	case Org:
		return "Org"
	case Folder:
		return "Folder"
	}
	return ""
}

func main() {
	var config struct {
		Data map[string]string `yaml:"data"`
	}

	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		hierarchy, err := processHierarchy(items)
		if err != nil {
			return items, err
		}
		if config.Data[renderFlagName] == "svg" {
			// if svg renderer, throw error if no outfile specified
			if config.Data[outputFlagName] == "" {
				return items, errors.New("output is a required argument")
			}
			out, err := os.OpenFile(config.Data[outputFlagName], os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				return items, err
			}
			return items, createDiagram(hierarchy, out)

		}
		// if renderer is not svg, outfile is optional
		// if outfile not specified print to stdout
		if config.Data[outputFlagName] == "" {
			return items, textTreeRenderer(hierarchy, os.Stderr)
		}
		// if outfile specified write to outfile
		out, err := os.OpenFile(config.Data[outputFlagName], os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return items, err
		}
		return items, textTreeRenderer(hierarchy, out)
	}

	fp := framework.SimpleProcessor{Filter: kio.FilterFunc(fn), Config: &config}
	cmd := command.Build(fp, command.StandaloneDisabled, false)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// gcpHierarchyResource represents a GCP folder or org placeholder
type gcpHierarchyResource struct {
	Name        string
	DisplayName string
	Parent      string
	ParentType  NodeType
	Type        NodeType
	Children    []*gcpHierarchyResource
}

// IsFolder checks if the object is a folder or an org placeholder
func (g *gcpHierarchyResource) IsFolder() bool {
	return g.Type == Folder
}

func gpcDrawSanitize(name string) string {
	val := strings.ReplaceAll(name, "-", "_")
	val = strings.ReplaceAll(val, ".", "_")
	return val
}

// GCPDrawName changes the name of the object to one that is supported by the
// GCPDraw DSL
func (g *gcpHierarchyResource) GCPDrawName() string {
	return gpcDrawSanitize(g.Name)
}

// GCPDrawParentName returns the name of the parent object and transforms it to
// a version supported by the GCP Draw DSL
func (g *gcpHierarchyResource) GCPDrawParentName() string {
	return gpcDrawSanitize(g.Parent)
}

func processHierarchy(items []*yaml.RNode) ([]*gcpHierarchyResource, error) {
	var hierarchy []*gcpHierarchyResource

	orgsAdded := make(map[string]struct{})
	hierarchyChildren := map[string][]*gcpHierarchyResource{}

	for _, item := range items {
		metadata, err := item.GetMeta()
		if err != nil {
			return nil, err
		}

		if metadata.Kind != folderKind || metadata.APIVersion != kccAPIVersion {
			return nil, errors.New("invalid resource kind or api version")
		}

		displayNameNode, err := item.Pipe(yaml.Lookup("spec", "displayName"))
		if err != nil {
			return nil, err
		}

		displayName, err := displayNameNode.String()
		if err != nil {
			return nil, err
		}

		if val, ok := metadata.Annotations[folderRefAnnotation]; ok {
			h := gcpHierarchyResource{
				Name:        metadata.Name,
				DisplayName: displayName,
				Parent:      val,
				ParentType:  Folder,
				Type:        Folder,
			}
			hierarchy = append(hierarchy, &h)
			hierarchyChildren[h.Parent] = append(hierarchyChildren[h.Parent], &h)
		} else if val, ok := metadata.Annotations[orgRefAnnotation]; ok {

			orgName := fmt.Sprintf("org-%s", val)
			h := gcpHierarchyResource{
				Name:        metadata.Name,
				DisplayName: displayName,
				Parent:      orgName,
				ParentType:  Org,
				Type:        Folder,
			}
			hierarchyChildren[h.Parent] = append(hierarchyChildren[h.Parent], &h)
			hierarchy = append(hierarchy, &h)

			if _, ok := orgsAdded[orgName]; !ok {
				orgsAdded[orgName] = struct{}{}
				hierarchy = append(hierarchy, &gcpHierarchyResource{
					Name:        orgName,
					Parent:      "",
					DisplayName: orgName,
					Type:        Org,
				})
			}
		}
	}
	for i := range hierarchy {
		if c, ok := hierarchyChildren[hierarchy[i].Name]; ok {
			hierarchy[i].Children = c
		}
	}
	return hierarchy, nil
}

// textTreeRenderer returns the tree visualization
func textTreeRenderer(hierarchy []*gcpHierarchyResource, output io.Writer) error {
	for _, h := range hierarchy {
		// find root org
		if h.Type == Org && h.Parent == "" {
			// add root org to top of tree viz
			_, writeErr := io.WriteString(output, fmt.Sprintf("\n%s\n", h.Name))

			if writeErr != nil {
				return writeErr
			}

			// generate rest of tree viz
			genErr := genTree(h, output, []bool{})

			if genErr != nil {
				return genErr
			}
		}
	}
	return nil
}

// genTree recursively generates the tree viz
func genTree(root *gcpHierarchyResource, output io.Writer, parentLastElems []bool) error {
	currChildren := root.Children
	for i, c := range currChildren {
		sep := ""
		// check if current element is last element, formatting for separator differs based on this
		isLastElem := i == len(currChildren)-1
		// parentLastElems is a bool slice which keeps track of depth and whether the parents are last elements or not
		for _, s := range parentLastElems {
			// if parent is last element fill with spaces, else use pipe + space
			// this is used for rendering cases like
			// | └─retail
			// |   ├─apps
			// |   └─data_and_analysis
			// where retail was a last element
			if s {
				sep += "  "
			} else {
				sep += "| "
			}
		}
		// if current child is the last element, use a different connector
		if isLastElem {
			sep += "└─"
		} else {
			sep += "├─"
		}
		// append to existing tree
		_, writeErr := io.WriteString(output, fmt.Sprintf("%s%s", sep, c.DisplayName))

		if writeErr != nil {
			return writeErr
		}

		// call genTree with current child as root
		genErr := genTree(c, output, append(parentLastElems, isLastElem))

		if genErr != nil {
			return genErr
		}

	}
	return nil
}

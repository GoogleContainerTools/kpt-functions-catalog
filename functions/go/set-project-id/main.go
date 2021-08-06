package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	kptv1 "github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/fix/v1"
)

var (
	kccAPIVersionRegex = regexp.MustCompile(`^([\w]+)\.cnrm\.cloud\.google\.com\/[\w]+$`)
)

const (
	projectIDSetterName = "project-id"
	projectIDAnnotation = "cnrm.cloud.google.com/project-id"
)

func findSetterNode(nodes []*yaml.RNode, path string) (*yaml.RNode, error) {
	for _, node := range nodes {
		np := node.GetAnnotations()[kioutil.PathAnnotation]
		if np == path {
			return node, nil
		}
	}
	return nil, fmt.Errorf(`file %s doesn't exist, please ensure the file specified in "configPath" exists and retry`, path)
}

func findKptfile(nodes []*yaml.RNode) (*kptv1.KptFile, error) {
	for _, node := range nodes {
		if node.GetAnnotations()[kioutil.PathAnnotation] == kptv1.KptFileName {
			kf, err := kptv1.ReadFile(node)
			if err != nil {
				return nil, fmt.Errorf("failed to read Kptfile: %v", err)
			}
			return kf, nil
		}
	}
	return nil, fmt.Errorf("unable to find Kptfile, please include --include-meta-resources flag if a Kptfile is present")
}

func setKptfile(nodes []*yaml.RNode, kf *kptv1.KptFile) error {
	b, err := yaml.Marshal(kf)
	if err != nil {
		return fmt.Errorf("failed to marshal updated Kptfile: %v", err)
	}
	kNode, err := yaml.Parse(string(b))
	if err != nil {
		return fmt.Errorf("failed to parse updated Kptfile: %v", err)
	}

	for i, _ := range nodes {
		if nodes[i].GetAnnotations()[kioutil.PathAnnotation] == kptv1.KptFileName {
			nodes[i] = kNode
			return nil
		}
	}
	return nil

}

func setSetters(nodes []*yaml.RNode, projectID string) error {
	kf, err := findKptfile(nodes)
	if err != nil {
		return fmt.Errorf("faild to find Kptfile: %v", err)
	}

	if kf.Pipeline != nil {
		for _, fn := range kf.Pipeline.Mutators {
			if !strings.Contains(fn.Image, "apply-setters") {
				continue
			}
			if fn.ConfigMap != nil {
				if fn.ConfigMap[projectIDSetterName] == "" {
					fn.ConfigMap[projectIDSetterName] = projectID
				}
				if err := setKptfile(nodes, kf); err != nil {
					return fmt.Errorf("failed to update Kptfile file: %v", err)
				}
				return nil
			} else if fn.ConfigPath != "" {
				settersConfig, err := findSetterNode(nodes, fn.ConfigPath)
				if err != nil {
					return fmt.Errorf("failed to find setter file: %v", err)
				}
				dataMap := settersConfig.GetDataMap()
				if dataMap[projectIDSetterName] == "" {
					dataMap[projectIDSetterName] = projectID
					settersConfig.SetDataMap(dataMap)
				}
				return nil
			} else {
				return fmt.Errorf("unable to find ConfigMap or ConfigPath fnConfig for apply-setters")
			}
		}
	} else {
		kf.Pipeline = &kptv1.Pipeline{}
	}

	fn := kptv1.Function{
		Image: "gcr.io/kpt-fn/apply-setters:v0.1",
		ConfigMap: map[string]string{
			projectIDSetterName: projectID,
		},
	}
	kf.Pipeline.Mutators = append(kf.Pipeline.Mutators, fn)
	if err := setKptfile(nodes, kf); err != nil {
		return fmt.Errorf("failed to update Kptfile file: %v", err)
	}

	return nil
}

func setProjectIDAnnotation(nodes []*yaml.RNode, projectID string) error {
	for _, node := range nodes {
		matches := kccAPIVersionRegex.FindStringSubmatch(node.GetApiVersion())
		if len(matches) == 2 && matches[1] != "core" {
			annotations := node.GetAnnotations()
			if _, ok := annotations[projectIDAnnotation]; !ok {
				annotations[projectIDAnnotation] = projectID
				if err := node.SetAnnotations(annotations); err != nil {
					return fmt.Errorf("failed to set project-id annotation: %v", err)
				}
				continue
			}
		}
	}
	return nil
}

type projectIDProcessor struct{}

func (p *projectIDProcessor) Process(resourceList *framework.ResourceList) error {
	dm := resourceList.FunctionConfig.GetDataMap()
	projectID := dm[projectIDSetterName]

	if err := setSetters(resourceList.Items, projectID); err != nil {
		return fmt.Errorf("failed to set project-id setter: %v", err)
	}
	if err := setProjectIDAnnotation(resourceList.Items, projectID); err != nil {
		return fmt.Errorf("failed to set project-id annotation: %v", err)
	}

	return nil
}

func main() {
	pp := projectIDProcessor{}
	cmd := command.Build(&pp, command.StandaloneEnabled, false)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

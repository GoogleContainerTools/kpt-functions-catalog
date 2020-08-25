package main

import (
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var namespace string
var allowClusterScoped bool

var warning = "due to the difficulty of tracking all the cluster-scoped resource kinds, especially cluster-scoped CRDs."

func main() {
	resourceList := &framework.ResourceList{}
	cmd := framework.Command(resourceList, func() error {
		// cmd.Execute() will parse the ResourceList.functionConfig into cmd.Flags from
		// the ResourceList.functionConfig.data field.
		clusterScopedResources := []string{}
		for i := range resourceList.Items {
			// modify the resources using the kyaml/yaml library:
			// https://pkg.go.dev/sigs.k8s.io/kustomize/kyaml/yaml
			node := resourceList.Items[i]
			kindNode, err := node.Pipe(yaml.Lookup("kind"))
			if err != nil {
				// ignore the node if it does not have a "kind" field
				continue
			}
			kind := yaml.GetValue(kindNode)
			typeMeta := yaml.TypeMeta{
				Kind: kind,
			}
			if typeMeta.IsNamespaceable() {
				// Set the metadata.namespace field
				if err := node.PipeE(yaml.LookupCreate(
					yaml.ScalarNode, "metadata", "namespace"),
					yaml.FieldSetter{StringValue: namespace}); err != nil {
					return err
				}
			} else {
				if !allowClusterScoped {
					apiVersionNode, err := node.Pipe(yaml.Lookup("apiVersion"))
					if err != nil {
						continue
					}
					apiVersion := yaml.GetValue(apiVersionNode)
					nameNode, err := node.Pipe(yaml.Lookup("metadata", "name"))
					if err != nil {
						continue
					}
					name := yaml.GetValue(nameNode)
					resID := fmt.Sprintf("apiVersion: %s, kind: %s, name: %s", apiVersion, kind, name)
					clusterScopedResources = append(clusterScopedResources, resID)
				}
			}
		}
		if len(clusterScopedResources) > 0 {
			return fmt.Errorf("the app config should only include namespace-scoped resources. "+
				"But the following cluster-scoped resources are found:\n%s\n", strings.Join(clusterScopedResources, "\n"))
		}
		return nil
	})

	cmd.Flags().BoolVar(&allowClusterScoped, "allow-cluster-scoped", true, "allow cluster-scoped resources or not. "+
		"This function may not be able to identify all the cluster-scoped resources "+warning)
	cmd.Flags().StringVar(&namespace, "namespace", "", "the namespace to be added into namespaced resources "+
		"This function may set the metadata.namespace field of some cluster-scoped resources "+warning)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

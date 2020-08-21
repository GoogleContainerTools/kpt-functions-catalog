package main

import (
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var namespace string

func main() {
	resourceList := &framework.ResourceList{}
	cmd := framework.Command(resourceList, func() error {
		// cmd.Execute() will parse the ResourceList.functionConfig into cmd.Flags from
		// the ResourceList.functionConfig.data field.
		for i := range resourceList.Items {
			// modify the resources using the kyaml/yaml library:
			// https://pkg.go.dev/sigs.k8s.io/kustomize/kyaml/yaml
			node := resourceList.Items[i]
			kindNode, err := node.Pipe(yaml.Lookup("kind"))
			if err != nil {
				// ignore the node if it does not have a "kind" field
				continue
			}
			typeMeta := yaml.TypeMeta{
				Kind: yaml.GetValue(kindNode),
			}
			if typeMeta.IsNamespaceable() {
				// Set the metadata.namespace field
				if err := node.PipeE(yaml.LookupCreate(
					yaml.ScalarNode, "metadata", "namespace"),
					yaml.FieldSetter{StringValue: namespace}); err != nil {
					return err
				}
			}
		}
		return nil
	})
	cmd.Flags().StringVar(&namespace, "namespace", "", "the namespace value")
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

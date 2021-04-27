package main

import (
	"fmt"
	"os"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/format/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
)

//nolint
func main() {
	resourceList := &framework.ResourceList{}

	cmd := framework.Command(resourceList, func() error {
		f := filters.FormatFilter{
			UseSchema: true,
		}
		_, err := f.Filter(resourceList.Items)
		if err != nil {
			return fmt.Errorf("failed to format resources: %w", err)
		}
		return nil
	})

	cmd.Short = generated.FormatShort
	cmd.Long = generated.FormatLong
	cmd.Example = generated.FormatExamples

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

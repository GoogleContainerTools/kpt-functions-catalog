package main

import (
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/fix/fixpkg"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/fix/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
)

//nolint
func main() {
	resourceList := &framework.ResourceList{}
	cmd := framework.Command(resourceList, func() error {
		s := &fixpkg.Fix{}
		var err error
		resourceList.Items, err = s.Filter(resourceList.Items)
		if err != nil {
			return fmt.Errorf("failed to fix resources: %w", err)
		}
		return nil
	})

	cmd.Short = generated.FixShort
	cmd.Long = generated.FixLong
	cmd.Example = generated.FixExamples

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}


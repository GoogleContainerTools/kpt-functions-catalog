package fnsdk_test

import (
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/thirdparty/kyaml/fnsdk"
)

// In this example, we read a field from the input object and print it to the log.

func Example_cSetField() {
	if err := fnsdk.AsMain(fnsdk.ResourceListProcessorFunc(setField)); err != nil {
		os.Exit(1)
	}
}

func setField(rl *fnsdk.ResourceList) error {
	for _, obj := range rl.Items {
		if obj.APIVersion() == "apps/v1" && obj.Kind() == "Deployment" {
			replicas := 10
			obj.SetOrDie(&replicas, "spec", "replicas")
		}
	}
	return nil
}

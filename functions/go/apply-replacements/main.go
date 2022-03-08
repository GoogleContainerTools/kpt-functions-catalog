package main

import (
	"os"

	sdk "github.com/GoogleContainerTools/kpt-functions-catalog/thirdparty/kyaml/fnsdk"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
)

func main() {
	if err := sdk.AsMain(sdk.ResourceListProcessorFunc(applyReplacements)); err != nil {
		os.Exit(1)
	}
}

func applyReplacements(rl *sdk.ResourceList) error {
	r := Replacements{}
	return r.Process(&framework.ResourceList{
		FunctionConfig: rl.FunctionConfig.ToRNode(),
		Items:          sdk.ToRNodes(rl.Items),
	})
}

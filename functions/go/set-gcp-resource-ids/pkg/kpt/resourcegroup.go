package common

import (
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

func IsResourceGroup(object *fn.KubeObject) bool {
	return object.IsGVK("kpt.dev", "v1alpha1", "ResourceGroup")
}

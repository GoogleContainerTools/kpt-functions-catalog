package custom

import (
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/api/filters/imagetag"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func SetAdditionalFieldSpec(image types.Image, objects fn.KubeObjects, additionalImageFields types.FsSlice, ctx *fn.Context) {
	for i, obj := range objects {
		objRN, err := yaml.Parse(obj.String())
		if err != nil {
			ctx.ResultInfo(err.Error(), obj)
		}
		filter := imagetag.Filter{
			ImageTag: image,
			FsSlice:  additionalImageFields,
		}

		filter.WithMutationTracker(mutationTracker(ctx, obj))
		err = filtersutil.ApplyToJSON(filter, objRN)
		if err != nil {
			ctx.ResultInfo(err.Error(), obj)
		}
		newObj, err := fn.ParseKubeObject([]byte(objRN.MustString()))
		if err != nil {
			ctx.ResultInfo(err.Error(), obj)
		}
		objects[i] = newObj
	}
}

func mutationTracker(ctx *fn.Context, ko *fn.KubeObject) func(key, value, tag string, node *yaml.RNode) {
	return func(key, value, tag string, node *yaml.RNode) {
		currentValue := node.YNode().Value
		msg := fmt.Sprintf("set image from %v to %v", currentValue, value)
		ctx.ResultInfo(msg, ko)
	}
}

package custom

import (
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-image/image_util"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/api/filters/imagetag"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// transformStruct transforms the struct inside image_util to the struct inside kustomize
func transformStruct(img image_util.Image, addImgFields image_util.FsSlice) (types.Image, types.FsSlice) {
	image := types.Image{
		Name:    img.Name,
		NewName: img.NewName,
		NewTag:  img.NewTag,
		Digest:  img.Digest,
	}
	additionalImageFields := types.FsSlice{}
	for _, v := range addImgFields {
		curFieldSpec := types.FieldSpec{
			Gvk: resid.Gvk{
				Group:   v.Gvk.Group,
				Version: v.Gvk.Version,
				Kind:    v.Gvk.Kind,
			},
			Path:               v.Path,
			CreateIfNotPresent: v.CreateIfNotPresent,
		}
		additionalImageFields = append(additionalImageFields, curFieldSpec)
	}
	return image, additionalImageFields
}

// SetAdditionalFieldSpec updates the image in user given fieldPaths. To be deprecated in around a year, to avoid possible invalid fieldPaths.
func SetAdditionalFieldSpec(img image_util.Image, objects fn.KubeObjects, addImgFields image_util.FsSlice, ctx *fn.Context) {
	image, additionalImageFields := transformStruct(img, addImgFields)

	for i, obj := range objects {
		objRN, err := yaml.Parse(obj.String())
		if err != nil {
			ctx.ResultErr(err.Error(), obj)
		}
		filter := imagetag.Filter{
			ImageTag: image,
			FsSlice:  additionalImageFields,
		}

		filter.WithMutationTracker(LogResultCallback(ctx, obj))
		err = filtersutil.ApplyToJSON(filter, objRN)
		if err != nil {
			ctx.ResultErr(err.Error(), obj)
		}
		newObj, err := fn.ParseKubeObject([]byte(objRN.MustString()))
		if err != nil {
			ctx.ResultErr(err.Error(), obj)
		}
		objects[i] = newObj
	}
}

func LogResultCallback(ctx *fn.Context, ko *fn.KubeObject) func(key, value, tag string, node *yaml.RNode) {
	return func(key, value, tag string, node *yaml.RNode) {
		currentValue := node.YNode().Value
		msg := fmt.Sprintf("updated image from %v to %v", currentValue, value)
		ctx.ResultInfo(msg, ko)
	}
}

package custom

import (
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/api/filters/imagetag"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// transformStruct transforms the struct inside third_party and transformer to the struct inside kustomize
func transformStruct(imgObj *fn.SubObject, addImgFields fn.SliceSubObjects) (types.Image, types.FsSlice) {
	image := types.Image{
		Name:    imgObj.NestedStringOrDie("name"),
		NewName: imgObj.NestedStringOrDie("newName"),
		NewTag:  imgObj.NestedStringOrDie("newTag"),
		Digest:  imgObj.NestedStringOrDie("digest"),
	}
	additionalImageFields := types.FsSlice{}
	for _, v := range addImgFields {
		curFieldSpec := types.FieldSpec{
			Gvk: resid.Gvk{
				Group:   v.NestedStringOrDie("group"),
				Version: v.NestedStringOrDie("version"),
				Kind:    v.NestedStringOrDie("kind"),
			},
			Path:               v.NestedStringOrDie("path"),
			CreateIfNotPresent: v.NestedBoolOrDie("create"),
		}
		additionalImageFields = append(additionalImageFields, curFieldSpec)
	}
	return image, additionalImageFields
}

// SetAdditionalFieldSpec updates the image in user given fieldPaths. To be deprecated in around a year, to avoid possible invalid fieldPaths.
func SetAdditionalFieldSpec(img *fn.SubObject, objects fn.KubeObjects, addImgFields fn.SliceSubObjects, ctx *fn.Context) {
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

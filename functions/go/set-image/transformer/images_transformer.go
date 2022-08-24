package transformer

import (
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-image/custom"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-image/image_util"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

// SetImage supports the set-image workflow, it uses Config to parse functionConfig, Transform to change the image
type SetImage struct {
	// Image is the desired image
	Image image_util.Image `json:"image,omitempty" yaml:"image,omitempty"`
	// ConfigMap keeps the data field that holds image information
	ConfigMap map[string]string `json:"data,omitempty" yaml:"data,omitempty"`
	// ONLY for kustomize, AdditionalImageFields is the user supplied fieldspec
	AdditionalImageFields image_util.FsSlice `json:"additionalImageFields,omitempty" yaml:"additionalImageFields,omitempty"`
	// context logs each detailed result
	context *fn.Context
	// resultCount logs the total count image change
	resultCount int
}

// Config transforms the data from ConfigMap to SetImage struct
func (t *SetImage) Config() error {
	for key, val := range t.ConfigMap {
		switch key {
		case "name":
			t.Image.Name = val
		case "newName":
			t.Image.NewName = val
		case "newTag":
			t.Image.NewTag = val
		case "digest":
			t.Image.Digest = val
		default:
			return fmt.Errorf("ConfigMap has wrong field name %v", key)
		}
	}
	return nil
}

// validateInput validates the inputs passed into via the functionConfig
func (t *SetImage) validateInput() error {
	// TODO: support container name and only one argument input in the next PR
	if t.Image.Name == "" {
		return fmt.Errorf("`name` field is missing from image selector")
	}
	if t.Image.NewName == "" && t.Image.NewTag == "" && t.Image.Digest == "" {
		return fmt.Errorf("must specify one of `newName`, `newTag`, or `digest`")
	}
	if t.Image.NewTag != "" && t.Image.Digest != "" {
		return fmt.Errorf("image `newTag` and `digest` both set, set only one")
	}
	return nil
}

func (t *SetImage) setPodSpecContainers(o *fn.KubeObject) error {
	podSpec := o.GetMap("spec").GetMap("template").GetMap("spec")
	for _, vecObj := range podSpec.GetSlice("containers") {
		if err := t.updateImages(vecObj, &t.Image, o); err != nil {
			return err
		}
	}
	for _, vecObj := range podSpec.GetSlice("iniContainers") {
		if err := t.updateImages(vecObj, &t.Image, o); err != nil {
			return err
		}
	}
	return nil
}

func (t *SetImage) setPodContainers(o *fn.KubeObject) error {
	spec := o.GetMap("spec")
	for _, vecObj := range spec.GetSlice("containers") {
		if err := t.updateImages(vecObj, &t.Image, o); err != nil {
			return err
		}
	}
	for _, vecObj := range spec.GetSlice("iniContainers") {
		if err := t.updateImages(vecObj, &t.Image, o); err != nil {
			return err
		}
	}
	return nil
}

func (t *SetImage) hasPodSpecContainers(o *fn.KubeObject) bool {
	spec := o.GetMap("spec")
	if spec == nil {
		return false
	}
	template := spec.GetMap("template")
	if template == nil {
		return false
	}
	podSpec := template.GetMap("spec")
	return podSpec != nil
}

func (t *SetImage) hasPodContainers(o *fn.KubeObject) bool {
	spec := o.GetMap("spec")
	return spec != nil
}

// getNewImageName return the new name for image field
func getNewImageName(oldValue string, newImage *image_util.Image) string {
	name, tag := image_util.Split(oldValue)
	if newImage.NewName != "" {
		name = newImage.NewName
	}
	if newImage.NewTag != "" {
		tag = ":" + newImage.NewTag
	}
	if newImage.Digest != "" {
		tag = "@" + newImage.Digest
	}
	newName := name + tag
	return newName
}

// updateImages update the image for a given fieldpath
func (t *SetImage) updateImages(o *fn.SubObject, newImage *image_util.Image, parentO *fn.KubeObject) error {
	oldValue := o.NestedStringOrDie("image")
	if !image_util.IsImageMatched(oldValue, newImage.Name) {
		return nil
	}
	newName := getNewImageName(oldValue, newImage)
	err := o.SetNestedString(newName, "image")

	msg := fmt.Sprintf("updated image from %v to %v", oldValue, newName)
	t.context.ResultInfo(msg, parentO)
	t.resultCount += 1
	return err
}

// Run implements the Runner interface that transforms the resource and log the results
func (t SetImage) Run(ctx *fn.Context, _ *fn.KubeObject, items fn.KubeObjects) {
	t.context = ctx
	err := t.Config()
	if err != nil {
		ctx.ResultErrAndDie(err.Error(), nil)
	}
	err = t.validateInput()
	if err != nil {
		ctx.ResultErrAndDie(err.Error(), nil)
	}

	for _, o := range items.Where(t.hasPodContainers) {
		err = t.setPodContainers(o)
		if err != nil {
			ctx.ResultErr(err.Error(), o)
		}
	}

	for _, o := range items.Where(t.hasPodSpecContainers) {
		err = t.setPodSpecContainers(o)
		if err != nil {
			ctx.ResultErr(err.Error(), o)
		}
	}

	if t.AdditionalImageFields != nil {
		custom.SetAdditionalFieldSpec(t.Image, items, t.AdditionalImageFields, t.context)
	}

	summary := fmt.Sprintf("summary: updated a total of %v image(s)", t.resultCount)
	ctx.ResultInfo(summary, nil)
}

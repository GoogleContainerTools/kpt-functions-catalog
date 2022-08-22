package transformer

import (
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-image/custom"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/api/image"
	"sigs.k8s.io/kustomize/api/types"
)

// SetImage supports the set-image workflow, it uses Config to parse functionConfig, Transform to change the image
type SetImage struct {
	// Image is the desired image
	Image types.Image `json:"image,omitempty" yaml:"image,omitempty"`
	// ConfigMap keeps the data field that holds image information
	ConfigMap map[string]string `json:"data,omitempty" yaml:"data,omitempty"`
	// ONLY for kustomize, AdditionalImageFields is the user supplied fieldspec
	AdditionalImageFields types.FsSlice `json:"additionalImageFields,omitempty" yaml:"additionalImageFields,omitempty"`
	// context logs each detailed result
	context *fn.Context
	// resultCount logs the total count image change
	resultCount int
}

// Config transforms the data from ConfigMap to SetImage struct
func (imageTrans *SetImage) Config() error {
	for key, val := range imageTrans.ConfigMap {
		if key == "name" {
			imageTrans.Image.Name = val
		} else if key == "newName" {
			imageTrans.Image.NewName = val
		} else if key == "newTag" {
			imageTrans.Image.NewTag = val
		} else if key == "digest" {
			imageTrans.Image.Digest = val
		} else {
			return fmt.Errorf("SubObject has unmatched field type: `data`")
		}
	}
	return nil
}

// validateInput validates the inputs passed into via the functionConfig
func (imageTrans *SetImage) validateInput() error {
	if imageTrans.Image.Name == "" {
		return fmt.Errorf("missing image name or typo")
	}
	if imageTrans.Image.NewName == "" && imageTrans.Image.NewTag == "" && imageTrans.Image.Digest == "" {
		return fmt.Errorf("missing image newName, newTag, or digest, could be typo or missing field")
	}
	if imageTrans.Image.NewTag != "" && imageTrans.Image.Digest != "" {
		return fmt.Errorf("image newTag and digest both set")
	}
	return nil
}

func (imageTrans *SetImage) setPodSpecContainers(o *fn.KubeObject) error {
	if spec := o.GetMap("spec"); spec != nil {
		if template := spec.GetMap("template"); template != nil {
			if podSpec := template.GetMap("spec"); podSpec != nil {
				for _, vecObj := range podSpec.GetSlice("containers") {
					if err := imageTrans.updateImages(vecObj, &imageTrans.Image, o); err != nil {
						return err
					}
				}
				for _, vecObj := range podSpec.GetSlice("iniContainers") {
					if err := imageTrans.updateImages(vecObj, &imageTrans.Image, o); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (imageTrans *SetImage) setPodContainers(o *fn.KubeObject) error {
	if spec := o.GetMap("spec"); spec != nil {
		for _, vecObj := range spec.GetSlice("containers") {
			if err := imageTrans.updateImages(vecObj, &imageTrans.Image, o); err != nil {
				return err
			}
		}
		for _, vecObj := range spec.GetSlice("iniContainers") {
			if err := imageTrans.updateImages(vecObj, &imageTrans.Image, o); err != nil {
				return err
			}
		}
	}
	return nil
}

func (imageTrans *SetImage) hasPodSpecContainers(o *fn.KubeObject) bool {
	if spec := o.GetMap("spec"); spec != nil {
		if template := spec.GetMap("template"); template != nil {
			if podSpec := template.GetMap("spec"); podSpec != nil {
				if spec.GetSlice("containers") != nil || spec.GetSlice("iniContainers") != nil {
					return true
				}
			}
		}
	}
	return false
}

func (imageTrans *SetImage) hasPodContainers(o *fn.KubeObject) bool {
	if spec := o.GetMap("spec"); spec != nil {
		if spec.GetSlice("containers") != nil || spec.GetSlice("iniContainers") != nil {
			return true
		}
	}
	return false
}

// getNewImageName return the new name for image field
func getNewImageName(oldValue string, newImage *types.Image) string {
	name, tag := image.Split(oldValue)
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
func (imageTrans *SetImage) updateImages(o *fn.SubObject, newImage *types.Image, parentO *fn.KubeObject) error {
	oldValue := o.NestedStringOrDie("image")
	if !image.IsImageMatched(oldValue, newImage.Name) {
		return nil
	}
	newName := getNewImageName(oldValue, newImage)
	err := o.SetNestedString(newName, "image")

	msg := fmt.Sprintf("set image from %v to %v", oldValue, newName)
	imageTrans.context.ResultInfo(msg, parentO)
	imageTrans.resultCount += 1
	return err
}

// Run implements the Runner interface that transforms the resource and log the results
func (si SetImage) Run(ctx *fn.Context, _ *fn.KubeObject, items fn.KubeObjects) {
	si.context = ctx
	err := si.Config()
	if err != nil {
		ctx.ResultErrAndDie(err.Error(), nil)
	}
	err = si.validateInput()
	if err != nil {
		ctx.ResultErrAndDie(err.Error(), nil)
	}

	for _, o := range items.Where(si.hasPodContainers) {
		err = si.setPodContainers(o)
		if err != nil {
			ctx.ResultErr(err.Error(), o)
		}
	}

	for _, o := range items.Where(si.hasPodSpecContainers) {
		err = si.setPodSpecContainers(o)
		if err != nil {
			ctx.ResultErr(err.Error(), o)
		}
	}

	if si.AdditionalImageFields != nil && len(si.AdditionalImageFields) != 0 {
		custom.SetAdditionalFieldSpec(si.Image, items, si.AdditionalImageFields, si.context)
	}

	summary := fmt.Sprintf("summary: total number of images updated %v", si.resultCount)
	ctx.ResultInfo(summary, nil)
}

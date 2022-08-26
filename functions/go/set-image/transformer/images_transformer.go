package transformer

import (
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-image/custom"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-image/third_party/sigs.k8s.io/kustomize/api/image"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-image/third_party/sigs.k8s.io/kustomize/api/types"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

// Image contains an image name, a new name, a new tag or digest, which will replace the original name and tag.
type Image struct {
	// Name is a tag-less image name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// NewName is the value used to replace the original name.
	NewName string `json:"newName,omitempty" yaml:"newName,omitempty"`

	// NewTag is the value used to replace the original tag.
	NewTag string `json:"newTag,omitempty" yaml:"newTag,omitempty"`

	// Digest is the value used to replace the original image tag.
	// If digest is present NewTag value is ignored.
	Digest string `json:"digest,omitempty" yaml:"digest,omitempty"`
}

// setImageResult keeps the old value and new value for logging
type setImageResult struct {
	currentValue  string
	proposedValue string
}

// SetImage supports the set-image workflow, it uses Config to parse functionConfig, Transform to change the image
type SetImage struct {
	// Image is the desired image
	Image Image `json:"image,omitempty" yaml:"image,omitempty"`
	// ConfigMap keeps the data field that holds image information
	DataFromDefaultConfig map[string]string `json:"data,omitempty" yaml:"data,omitempty"`
	// ONLY for kustomize, AdditionalImageFields is the user supplied fieldspec
	AdditionalImageFields types.FsSlice `json:"additionalImageFields,omitempty" yaml:"additionalImageFields,omitempty"`
	// resultCount logs the total count image change
	resultCount int
}

// Run implements the Runner interface that transforms the resource and log the results
func (t SetImage) Run(ctx *fn.Context, functionConfig *fn.KubeObject, items fn.KubeObjects) {
	err := t.configDefaultData()
	if err != nil {
		ctx.ResultErrAndDie(err.Error(), nil)
	}
	err = t.validateInput()
	if err != nil {
		ctx.ResultErrAndDie(err.Error(), nil)
	}

	for _, o := range items.Where(t.hasPodContainers) {
		err, result := t.setPodContainers(o)
		t.logResult(ctx, err, result, o)
	}

	for _, o := range items.Where(t.hasPodSpecContainers) {
		err, result := t.setPodSpecContainers(o)
		t.logResult(ctx, err, result, o)
	}

	if t.AdditionalImageFields != nil {
		custom.SetAdditionalFieldSpec(functionConfig.GetMap("image"), items, functionConfig.GetSlice("additionalImageFields"), ctx)
	}

	summary := fmt.Sprintf("summary: updated a total of %v image(s)", t.resultCount)
	ctx.ResultInfo(summary, nil)
}

// Config transforms the data from ConfigMap to SetImage struct
func (t *SetImage) configDefaultData() error {
	for key, val := range t.DataFromDefaultConfig {
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
		return fmt.Errorf("must specify `name`")
	}
	if t.Image.NewName == "" && t.Image.NewTag == "" && t.Image.Digest == "" {
		return fmt.Errorf("must specify one of `newName`, `newTag`, or `digest`")
	}
	if t.Image.NewTag != "" && t.Image.Digest != "" {
		return fmt.Errorf("image `newTag` and `digest` both set, set only one")
	}
	return nil
}

// updateContainerImages updates the images inside containers, return potential error, and a list of logging results
func (t *SetImage) updateContainerImages(pod *fn.SubObject) (error, []setImageResult) {
	var containers fn.SliceSubObjects
	containers = append(containers, pod.GetSlice("iniContainers")...)
	containers = append(containers, pod.GetSlice("containers")...)

	var result []setImageResult
	for _, o := range containers {
		oldValue := o.NestedStringOrDie("image")
		if !image.IsImageMatched(oldValue, t.Image.Name) {
			continue
		}
		newName := getNewImageName(oldValue, t.Image)
		if oldValue == newName {
			return nil, nil
		}

		if err := o.SetNestedString(newName, "image"); err != nil {
			return err, nil
		}
		t.resultCount += 1

		result = append(result, setImageResult{
			currentValue:  oldValue,
			proposedValue: newName,
		})
	}
	return nil, result
}

func (t *SetImage) setPodSpecContainers(o *fn.KubeObject) (error, []setImageResult) {
	spec := o.GetMap("spec")
	if spec == nil {
		return nil, nil
	}
	template := spec.GetMap("template")
	if template == nil {
		return nil, nil
	}
	podSpec := template.GetMap("spec")
	err, result := t.updateContainerImages(podSpec)
	if err != nil {
		return err, nil
	}
	return nil, result
}

func (t *SetImage) setPodContainers(o *fn.KubeObject) (error, []setImageResult) {
	spec := o.GetMap("spec")
	if spec == nil {
		return nil, nil
	}
	err, result := t.updateContainerImages(spec)
	if err != nil {
		return err, nil
	}
	return nil, result
}

func (t *SetImage) hasPodSpecContainers(o *fn.KubeObject) bool {
	return o.IsGVK("", "", "PodTemplate")
}

func (t *SetImage) hasPodContainers(o *fn.KubeObject) bool {
	return o.IsGVK("", "", "Pod")
}

// getNewImageName return the new name for image field
func getNewImageName(oldValue string, newImage Image) string {
	name, tag, digest := image.Split(oldValue)
	if newImage.NewName != "" {
		name = newImage.NewName
	}
	if newImage.NewTag != "" {
		tag = ":" + newImage.NewTag
	}
	if newImage.Digest != "" {
		tag = "@" + newImage.Digest
	}
	var newName string
	if tag == "" {
		newName = name + digest
	} else {
		newName = name + tag
	}

	return newName
}

func (t SetImage) logResult(ctx *fn.Context, err error, result []setImageResult, o *fn.KubeObject) {
	if err != nil {
		ctx.ResultErr(err.Error(), o)
	}
	for _, val := range result {
		msg := fmt.Sprintf("updated image from %v to %v", val.currentValue, val.proposedValue)
		ctx.ResultInfo(msg, o)
	}
}

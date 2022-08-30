package transformer

import (
	"fmt"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-image/custom"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-image/third_party/sigs.k8s.io/kustomize/api/image"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-image/third_party/sigs.k8s.io/kustomize/api/types"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

type Image struct {
	// DEPRECATED
	//Name is a tag-less image name. should be deprecate, means image name
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// ContainerName is the name for container
	ContainerName string `json:"containerName,omitempty" yaml:"containerName,omitempty"`

	// ImageName is the image name
	ImageName string `json:"imageName,omitempty" yaml:"imageName,omitempty"`

	// NewName is the value used to replace the original name, replace image name
	NewName string `json:"newName,omitempty" yaml:"newName,omitempty"`

	// NewTag is the value used to replace the original tag.
	NewTag string `json:"newTag,omitempty" yaml:"newTag,omitempty"`

	// Digest is the value used to replace the original image tag.
	// If digest is present NewTag value is ignored.
	Digest string `json:"digest,omitempty" yaml:"digest,omitempty"`
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
	// currentImageNames records current image name, if no image name is provided as input, there should only be one ImageNames
	currentImageName string
	// isImageSelected checks if the input provide the image name or not
	isImageSelected bool
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

	for _, o := range items {
		switch o.GetKind() {
		case "Pod":
			errList := t.setPodContainers(o)
			for _, e := range errList {
				if strings.Contains(e.Error(), "must") {
					ctx.ResultErrAndDie(e.Error(), o)
				}
				ctx.ResultErr(e.Error(), o)
			}
		case "Deployment", "StatefulSet", "ReplicaSet", "DaemonSet", "PodTemplate":
			errList := t.setPodSpecContainers(o)
			for _, e := range errList {
				if strings.Contains(e.Error(), "must") {
					ctx.ResultErrAndDie(e.Error(), o)
				}
				ctx.ResultErr(e.Error(), o)
			}
		}
	}

	if t.AdditionalImageFields != nil {
		custom.SetAdditionalFieldSpec(functionConfig.GetMap("image"), items, functionConfig.GetSlice("additionalImageFields"), ctx, &t.resultCount)
	}

	summary := fmt.Sprintf("summary: updated a total of %v image(s)", t.resultCount)
	ctx.ResultInfo(summary, nil)
}

// configDefaultData transforms the data from ConfigMap to SetImage struct
func (t *SetImage) configDefaultData() error {
	for key, val := range t.DataFromDefaultConfig {
		switch key {
		case "name":
			t.Image.ImageName = val
		case "containerName":
			t.Image.ContainerName = val
		case "imageName":
			t.Image.ImageName = val
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
	// if user does not input image name or container name, there should be only one image name to select from
	if t.Image.Name == "" && t.Image.ImageName == "" && t.Image.ContainerName == "" {
		t.isImageSelected = false
	} else {
		t.isImageSelected = true
	}

	if t.Image.Name != "" && t.Image.ImageName != "" && t.Image.Name != t.Image.ImageName {
		return fmt.Errorf("must not fill `imageName` and `name` at same time, their values should be equal")
	}
	if t.Image.NewName == "" && t.Image.NewTag == "" && t.Image.Digest == "" {
		return fmt.Errorf("must specify one of `newName`, `newTag`, or `digest`")
	}
	return nil
}

func matchImage(oldName string, oldContainer string, newImage *Image) (bool, error) {
	// name would be deprecated, means image name
	if newImage.Name != "" {
		newImage.ImageName = newImage.Name
	}

	// match image name, no container name
	if newImage.ImageName != "" && newImage.ContainerName == "" {
		if !image.IsImageMatched(oldName, newImage.ImageName) {
			return false, nil
		}
	}
	// match container name, no image name
	if newImage.ImageName == "" && newImage.ContainerName != "" {
		if oldContainer != newImage.ContainerName {
			return false, nil
		}
	}
	// match both
	if newImage.ImageName != "" && newImage.ContainerName != "" {
		if !image.IsImageMatched(oldName, newImage.ImageName) {
			return false, nil
		}
		if oldContainer != newImage.ContainerName {
			msg := fmt.Sprintf("container name `%v` does not match `%v`, only image name matches", newImage.ContainerName, oldContainer)
			warning := fn.Result{
				Message:  msg,
				Severity: fn.Warning,
			}
			return true, warning
		}
	}
	return true, nil
}

func (t *SetImage) isImageNameUnique(oldValue string) bool {
	name, _, _ := image.Split(oldValue)
	if t.currentImageName == "" {
		t.currentImageName = name
	} else {
		return t.currentImageName == name
	}
	return true
}

// updateContainerImages updates the images inside containers, return potential error
func (t *SetImage) updateContainerImages(pod *fn.SubObject) []error {
	var containers fn.SliceSubObjects
	containers = append(containers, pod.GetSlice("iniContainers")...)
	containers = append(containers, pod.GetSlice("containers")...)

	var warningList []error
	for _, o := range containers {
		oldValue := o.NestedStringOrDie("image")
		oldContainer := o.NestedStringOrDie("name")

		if t.isImageSelected == false {
			if !t.isImageNameUnique(oldValue) {
				msg := fmt.Sprintf("must specify `imageName`, resources contain non-unique image names")
				err := fn.Result{
					Message:  msg,
					Severity: fn.Error,
				}
				return []error{err}
			}
		}

		matched, warning := matchImage(oldValue, oldContainer, &t.Image)
		if !matched {
			continue
		}

		newName := getNewImageName(oldValue, t.Image)
		if oldValue == newName {
			continue
		}

		if warning != nil {
			warningList = append(warningList, warning)
		}

		if err := o.SetNestedString(newName, "image"); err != nil {
			return []error{err}
		}
		t.resultCount += 1
	}
	return warningList
}

func (t *SetImage) setPodSpecContainers(o *fn.KubeObject) []error {
	spec := o.GetMap("spec")
	if spec == nil {
		return nil
	}
	template := spec.GetMap("template")
	if template == nil {
		return nil
	}
	podSpec := template.GetMap("spec")
	err := t.updateContainerImages(podSpec)
	if err != nil {
		return err
	}
	return nil
}

func (t *SetImage) setPodContainers(o *fn.KubeObject) []error {
	spec := o.GetMap("spec")
	if spec == nil {
		return nil
	}
	err := t.updateContainerImages(spec)
	if err != nil {
		return err
	}
	return nil
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

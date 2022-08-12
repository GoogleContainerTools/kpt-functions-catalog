package transformer

import (
	"fmt"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/api/image"
	"sigs.k8s.io/kustomize/api/types"
)

const FnConfigKind = "SetImage"

type FieldPath []string

type Image struct {
	// Name is a tag-less image name. should be deprecate, means image name
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

// ImageTransformer supports the set-image workflow, it uses Config to parse functionConfig, Transform to change the image
type ImageTransformer struct {
	// Image is the desired image
	Image Image
	// AdditionalImageFields is used to specify additional fields to set image
	AdditionalImageFields types.FsSlice
	// Results logs the changes to the KRM resource image
	Results fn.Results
	// ResultCount logs the total count image change
	ResultCount int
}

// Config parse the functionConfig kubeObject to the fields in the ImageTransformer
func (imageTrans *ImageTransformer) Config(functionConfig *fn.KubeObject) error {
	switch {
	case functionConfig.IsEmpty():
		return fmt.Errorf("Config is Empty, failed to configure function: `functionConfig` must be either a `ConfigMap` or `SetImage`")
	case functionConfig.IsGVK("", "v1", "ConfigMap"):
		functionConfig.GetOrDie(&imageTrans.Image, "data")
	case functionConfig.IsGVK(fn.KptFunctionGroup, fn.KptFunctionVersion, FnConfigKind):
		functionConfig.GetOrDie(&imageTrans.Image, "image")
		functionConfig.GetOrDie(&imageTrans.AdditionalImageFields, "additionalImageFields")
	default:
		return fmt.Errorf("unknown functionConfig Kind=%v ApiVersion=%v, expect `%v` or `ConfigMap` with correct formatting",
			functionConfig.GetKind(), functionConfig.GetAPIVersion(), FnConfigKind)
	}
	if err := imageTrans.validateInput(); err != nil {
		return err
	}
	return nil
}

// validateInput validates the inputs passed into via the functionConfig
func (imageTrans *ImageTransformer) validateInput() error {
	if imageTrans.Image.Name == "" && imageTrans.Image.ContainerName == "" && imageTrans.Image.ImageName == "" {
		return fmt.Errorf("missing image name or container name")
	}
	if imageTrans.Image.NewName == "" && imageTrans.Image.NewTag == "" && imageTrans.Image.Digest == "" {
		return fmt.Errorf("missing image newName, newTag, or digest")
	}
	if imageTrans.Image.NewTag != "" && imageTrans.Image.Digest != "" {
		return fmt.Errorf("image newTag and digest both set")
	}
	return nil
}

// Transform updates the image in pod
func (imageTrans *ImageTransformer) Transform(objects fn.KubeObjects) error {
	// using unit test and pass in empty string would provide a nil; an empty file in e2e would provide 0 object
	if objects.Len() == 0 || objects[0] == nil {
		newResult := fn.GeneralResult("no input resources", fn.Info)
		imageTrans.Results = append(imageTrans.Results, newResult)
		return nil
	}
	for _, o := range objects.WhereNot(func(o *fn.KubeObject) bool { return o.IsLocalConfig() }) {
		imageTrans.setPodContainers(o)
		imageTrans.setPodSpecContainers(o)
	}
	return nil
}

func (imageTrans *ImageTransformer) addWarning(o *fn.KubeObject) {
	warning := &fn.Result{
		Message:  "container name does not match, only image name matches",
		Severity: fn.Warning,
		ResourceRef: &fn.ResourceRef{
			APIVersion: o.GetAPIVersion(),
			Kind:       o.GetKind(),
			Name:       o.GetName(),
			Namespace:  o.GetNamespace(),
		},
		File: &fn.File{
			Path:  o.PathAnnotation(),
			Index: o.IndexAnnotation(),
		},
		Tags: nil,
	}
	imageTrans.Results = append(imageTrans.Results, warning)
}

func (imageTrans *ImageTransformer) setPodSpecContainers(o *fn.KubeObject) {
	if spec := o.GetMap("spec"); spec != nil {
		if template := spec.GetMap("template"); template != nil {
			if podSpec := template.GetMap("spec"); podSpec != nil {
				for _, vecObj := range podSpec.GetSlice("containers") {
					fieldPath := FieldPath{"spec", "template", "spec", "containers"}
					imageTrans.updateImages(vecObj, &imageTrans.Image, o, fieldPath)
				}
				for _, vecObj := range podSpec.GetSlice("iniContainers") {
					fieldPath := FieldPath{"spec", "template", "spec", "iniContainers"}
					imageTrans.updateImages(vecObj, &imageTrans.Image, o, fieldPath)
				}
			}
		}
	}
}

func (imageTrans *ImageTransformer) setPodContainers(o *fn.KubeObject) {
	if spec := o.GetMap("spec"); spec != nil {
		for _, vecObj := range spec.GetSlice("containers") {
			fieldPath := FieldPath{"spec", "containers"}
			imageTrans.updateImages(vecObj, &imageTrans.Image, o, fieldPath)
		}
		for _, vecObj := range spec.GetSlice("iniContainers") {
			fieldPath := FieldPath{"spec", "iniContainers"}
			imageTrans.updateImages(vecObj, &imageTrans.Image, o, fieldPath)
		}
	}
}

func getNewImageName(oldValue string, newImage *Image) string {
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

func matchImage(oldName string, oldContainer string, newImage *Image) (bool, error) {
	// name would be deprecated later, means image name
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
			warning := fn.Result{
				Message:  "container name does not match, only image name matches",
				Severity: fn.Warning,
			}
			return true, warning
		}
	}
	return true, nil
}

// updateImages the update process for each image, if error happened, it panics
func (imageTrans *ImageTransformer) updateImages(o *fn.SubObject, newImage *Image, parentO *fn.KubeObject, fieldPath FieldPath) {
	oldValue := o.NestedStringOrDie("image")
	container := o.NestedStringOrDie("name")

	matched, err := matchImage(oldValue, container, newImage)
	if matched {
		newName := getNewImageName(oldValue, newImage)
		o.SetNestedStringOrDie(newName, "image")
		fieldPath = append(fieldPath, "image")
		imageTrans.LogResult(parentO, oldValue, newName, fieldPath)
		imageTrans.ResultCount += 1
	}
	if err != nil {
		imageTrans.addWarning(parentO)
	}

}

func (imageTrans *ImageTransformer) LogResult(o *fn.KubeObject, oldValue string, newValue string, fieldPath FieldPath) {
	newResult := fn.Result{
		Message:  fmt.Sprintf("set image from %v to %v", oldValue, newValue),
		Severity: fn.Info,
		ResourceRef: &fn.ResourceRef{
			APIVersion: o.GetAPIVersion(),
			Kind:       o.GetKind(),
			Name:       o.GetName(),
			Namespace:  o.GetNamespace(),
		},
		Field: &fn.Field{
			Path:          strings.Join(fieldPath, "."),
			CurrentValue:  oldValue,
			ProposedValue: newValue,
		},
		File: &fn.File{
			Path:  o.PathAnnotation(),
			Index: o.IndexAnnotation(),
		},
		Tags: nil,
	}
	imageTrans.Results = append(imageTrans.Results, &newResult)
}

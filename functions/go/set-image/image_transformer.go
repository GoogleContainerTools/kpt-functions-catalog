package main

import (
	"fmt"

	sdk "github.com/GoogleContainerTools/kpt-functions-catalog/thirdparty/kyaml/fnsdk"
	"sigs.k8s.io/kustomize/api/filters/imagetag"
	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	fnConfigGroup      = "fn.kpt.dev"
	fnConfigVersion    = "v1alpha1"
	fnConfigAPIVersion = fnConfigGroup + "/" + fnConfigVersion
	fnConfigKind       = "SetImage"
)

// getDefaultImageFields returns default image FieldSpecs
func getDefaultImageFields() (types.FsSlice, error) {
	type defaultConfig struct {
		FieldSpecs types.FsSlice `json:"images,omitempty" yaml:"images,omitempty"`
	}
	defaultConfigString := builtinpluginconsts.GetDefaultFieldSpecsAsMap()["images"]
	var tc defaultConfig
	err := yaml.Unmarshal([]byte(defaultConfigString), &tc)
	return tc.FieldSpecs, err
}

// validGVK returns whether the given sdk.KubeObject is of the given apiVersion/kind
func validGVK(ko *sdk.KubeObject, apiVersion, kind string) bool {
	return ko.APIVersion() == apiVersion && ko.Kind() == kind
}

type SetImage struct {
	// Image Desired image tag.
	Image types.Image `json:"image,omitempty" yaml:"image,omitempty"`
	// AdditionalImageFields is used to specify additional fields to set image.
	AdditionalImageFields types.FsSlice `json:"additionalImageFields,omitempty" yaml:"additionalImageFields,omitempty"`
}

// validateInput validates the inputs passed into via the functionConfig
func (si *SetImage) validateInput() error {
	if si.Image.Name == "" {
		return fmt.Errorf("missing image name")
	}
	if si.Image.NewName == "" && si.Image.NewTag == "" && si.Image.Digest == "" {
		return fmt.Errorf("missing image newName, newTag, or digest")
	}
	if si.Image.NewTag != "" && si.Image.Digest != "" {
		return fmt.Errorf("image newTag and digest both set")
	}
	return nil
}

// Config initializes SetImage from a functionConfig sdk.KubeObject
func (si *SetImage) Config(functionConfig *sdk.KubeObject) error {
	si.Image = types.Image{}
	si.AdditionalImageFields = nil
	switch {
	case validGVK(functionConfig, "v1", "ConfigMap"):
		if found, err := functionConfig.Get(&si.Image, "data"); err != nil {
			return fmt.Errorf("unable to convert functionConfig to v1 ConfigMap:\n%w", err)
		} else if !found {
			return fmt.Errorf("unable to get field data from functionConfig")
		}
	case validGVK(functionConfig, fnConfigAPIVersion, fnConfigKind):
		if err := functionConfig.As(si); err != nil {
			return fmt.Errorf("unable to convert functionConfig to %s %s:\n%w",
				fnConfigAPIVersion, fnConfigKind, err)
		}
	default:
		return fmt.Errorf("`functionConfig` must be a `ConfigMap` or `%s`", fnConfigKind)
	}
	if err := si.validateInput(); err != nil {
		return err
	}
	defaultImageFields, err := getDefaultImageFields()
	if err != nil {
		return err
	}
	si.AdditionalImageFields = append(si.AdditionalImageFields, defaultImageFields...)
	return nil
}

// Transform set image out of place and returns a new []*sdk.KubeObject
func (si *SetImage) Transform(items []*sdk.KubeObject) ([]*sdk.KubeObject, error) {
	var transformedItems []*sdk.KubeObject
	for _, obj := range items {
		objRN := obj.ToRNode()
		err := filtersutil.ApplyToJSON(imagetag.Filter{
			ImageTag:  si.Image,
			FsSlice:   si.AdditionalImageFields,
		}, objRN)
		if err != nil {
			return nil, err
		}
		transformedItems = append(transformedItems, sdk.NewFromRNode(objRN))
	}
	return transformedItems, nil
}

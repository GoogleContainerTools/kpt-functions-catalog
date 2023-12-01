package main

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestSortResultItems(t *testing.T) {
	testcases := []struct {
		name   string
		input  framework.Results
		output framework.Results
	}{
		{
			name: "sort based on severity",
			input: framework.Results{
				&framework.Result{
					Message:  "Error message 1",
					Severity: framework.Info,
					File:     &framework.File{},
				},
				&framework.Result{
					Message:  "Error message 2",
					Severity: framework.Error,
					File:     &framework.File{},
				},
			},
			output: framework.Results{
				&framework.Result{
					Message:  "Error message 2",
					Severity: framework.Error,
					File:     &framework.File{},
				},
				&framework.Result{
					Message:  "Error message 1",
					Severity: framework.Info,
					File:     &framework.File{},
				},
			},
		},
		{
			name: "sort based on file",
			input: framework.Results{
				&framework.Result{
					Message:  "Error message",
					Severity: framework.Error,
					File: &framework.File{
						Path:  "resource.yaml",
						Index: 1,
					},
				},
				&framework.Result{
					Message:  "Error message",
					Severity: framework.Info,
					File: &framework.File{
						Path:  "resource.yaml",
						Index: 0,
					},
				},
				&framework.Result{
					Message:  "Error message",
					Severity: framework.Info,
					File: &framework.File{
						Path:  "other-resource.yaml",
						Index: 0,
					},
				},
				&framework.Result{
					Message:  "Error message",
					Severity: framework.Warning,
					File: &framework.File{
						Path:  "resource.yaml",
						Index: 2,
					},
				},
			},
			output: framework.Results{
				&framework.Result{
					Message:  "Error message",
					Severity: framework.Info,
					File: &framework.File{
						Path:  "other-resource.yaml",
						Index: 0,
					},
				},
				&framework.Result{
					Message:  "Error message",
					Severity: framework.Info,
					File: &framework.File{
						Path:  "resource.yaml",
						Index: 0,
					},
				},
				&framework.Result{
					Message:  "Error message",
					Severity: framework.Error,
					File: &framework.File{
						Path:  "resource.yaml",
						Index: 1,
					},
				},
				&framework.Result{
					Message:  "Error message",
					Severity: framework.Warning,
					File: &framework.File{
						Path:  "resource.yaml",
						Index: 2,
					},
				},
			},
		},
		{
			name: "sort based on other fields",
			input: framework.Results{
				&framework.Result{
					Message:  "Error message",
					Severity: framework.Error,
					File:     &framework.File{},
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: &framework.Field{
						Path: "spec",
					},
				},
				&framework.Result{
					Message:  "Error message",
					Severity: framework.Error,
					File:     &framework.File{},
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: &framework.Field{
						Path: "metadata.name",
					},
				},
				&framework.Result{
					Message:  "Another error message",
					Severity: framework.Error,
					File:     &framework.File{},
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: &framework.Field{
						Path: "metadata.name",
					},
				},
				&framework.Result{
					Message:  "Another error message",
					Severity: framework.Error,
					File:     &framework.File{},
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "ConfigMap",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: &framework.Field{
						Path: "metadata.name",
					},
				},
			},
			output: framework.Results{
				&framework.Result{
					Message:  "Another error message",
					Severity: framework.Error,
					File:     &framework.File{},
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "ConfigMap",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: &framework.Field{
						Path: "metadata.name",
					},
				},
				&framework.Result{
					Message:  "Another error message",
					Severity: framework.Error,
					File:     &framework.File{},
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: &framework.Field{
						Path: "metadata.name",
					},
				},
				&framework.Result{
					Message:  "Error message",
					Severity: framework.Error,
					File:     &framework.File{},
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: &framework.Field{
						Path: "metadata.name",
					},
				},
				&framework.Result{
					Message:  "Error message",
					Severity: framework.Error,
					File:     &framework.File{},
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: &framework.Field{
						Path: "spec",
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		sortResultItems(tc.input)
		if !reflect.DeepEqual(tc.input, tc.output) {
			t.Errorf("in testcase %q, expect: %#v, but got: %#v", tc.name, tc.output, tc.input)
		}
	}
}

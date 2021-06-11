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
		input  []framework.ResultItem
		output []framework.ResultItem
	}{
		{
			name: "sort based on severity",
			input: []framework.ResultItem{
				{
					Message:  "Error message 1",
					Severity: framework.Info,
				},
				{
					Message:  "Error message 2",
					Severity: framework.Error,
				},
			},
			output: []framework.ResultItem{
				{
					Message:  "Error message 2",
					Severity: framework.Error,
				},
				{
					Message:  "Error message 1",
					Severity: framework.Info,
				},
			},
		},
		{
			name: "sort based on file",
			input: []framework.ResultItem{
				{
					Message:  "Error message",
					Severity: framework.Error,
					File: framework.File{
						Path:  "resource.yaml",
						Index: 1,
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Info,
					File: framework.File{
						Path:  "resource.yaml",
						Index: 0,
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Info,
					File: framework.File{
						Path:  "other-resource.yaml",
						Index: 0,
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Warning,
					File: framework.File{
						Path:  "resource.yaml",
						Index: 2,
					},
				},
			},
			output: []framework.ResultItem{
				{
					Message:  "Error message",
					Severity: framework.Info,
					File: framework.File{
						Path:  "other-resource.yaml",
						Index: 0,
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Info,
					File: framework.File{
						Path:  "resource.yaml",
						Index: 0,
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Error,
					File: framework.File{
						Path:  "resource.yaml",
						Index: 1,
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Warning,
					File: framework.File{
						Path:  "resource.yaml",
						Index: 2,
					},
				},
			},
		},

		{
			name: "sort based on other fields",
			input: []framework.ResultItem{
				{
					Message:  "Error message",
					Severity: framework.Error,
					ResourceRef: yaml.ResourceMeta{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						ObjectMeta: yaml.ObjectMeta{
							NameMeta: yaml.NameMeta{
								Namespace: "foo-ns",
								Name:      "bar",
							},
						},
					},
					Field: framework.Field{
						Path: "spec",
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Error,
					ResourceRef: yaml.ResourceMeta{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						ObjectMeta: yaml.ObjectMeta{
							NameMeta: yaml.NameMeta{
								Namespace: "foo-ns",
								Name:      "bar",
							},
						},
					},
					Field: framework.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Another error message",
					Severity: framework.Error,
					ResourceRef: yaml.ResourceMeta{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						ObjectMeta: yaml.ObjectMeta{
							NameMeta: yaml.NameMeta{
								Namespace: "foo-ns",
								Name:      "bar",
							},
						},
					},
					Field: framework.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Another error message",
					Severity: framework.Error,
					ResourceRef: yaml.ResourceMeta{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "ConfigMap",
						},
						ObjectMeta: yaml.ObjectMeta{
							NameMeta: yaml.NameMeta{
								Namespace: "foo-ns",
								Name:      "bar",
							},
						},
					},
					Field: framework.Field{
						Path: "metadata.name",
					},
				},
			},
			output: []framework.ResultItem{
				{
					Message:  "Another error message",
					Severity: framework.Error,
					ResourceRef: yaml.ResourceMeta{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "ConfigMap",
						},
						ObjectMeta: yaml.ObjectMeta{
							NameMeta: yaml.NameMeta{
								Namespace: "foo-ns",
								Name:      "bar",
							},
						},
					},
					Field: framework.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Another error message",
					Severity: framework.Error,
					ResourceRef: yaml.ResourceMeta{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						ObjectMeta: yaml.ObjectMeta{
							NameMeta: yaml.NameMeta{
								Namespace: "foo-ns",
								Name:      "bar",
							},
						},
					},
					Field: framework.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Error,
					ResourceRef: yaml.ResourceMeta{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						ObjectMeta: yaml.ObjectMeta{
							NameMeta: yaml.NameMeta{
								Namespace: "foo-ns",
								Name:      "bar",
							},
						},
					},
					Field: framework.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Error,
					ResourceRef: yaml.ResourceMeta{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						ObjectMeta: yaml.ObjectMeta{
							NameMeta: yaml.NameMeta{
								Namespace: "foo-ns",
								Name:      "bar",
							},
						},
					},
					Field: framework.Field{
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

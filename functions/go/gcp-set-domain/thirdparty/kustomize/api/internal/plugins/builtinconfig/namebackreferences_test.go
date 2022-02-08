// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package builtinconfig

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
)

func TestMergeAll(t *testing.T) {
	fsSlice1 := []types.FieldSpec{
		{
			Gvk: resid.Gvk{
				Kind: "Pod",
			},
			Path:               "path/to/a/name",
			CreateIfNotPresent: false,
		},
		{
			Gvk: resid.Gvk{
				Kind: "Deployment",
			},
			Path:               "another/path/to/some/name",
			CreateIfNotPresent: false,
		},
	}
	fsSlice2 := []types.FieldSpec{
		{
			Gvk: resid.Gvk{
				Kind: "Job",
			},
			Path:               "morepath/to/name",
			CreateIfNotPresent: false,
		},
		{
			Gvk: resid.Gvk{
				Kind: "StatefulSet",
			},
			Path:               "yet/another/path/to/a/name",
			CreateIfNotPresent: false,
		},
	}

	nbrsSlice1 := nbrSlice{
		{
			Gvk: resid.Gvk{
				Kind: "ConfigMap",
			},
			Referrers: fsSlice1,
		},
		{
			Gvk: resid.Gvk{
				Kind: "Secret",
			},
			Referrers: fsSlice2,
		},
	}
	nbrsSlice2 := nbrSlice{
		{
			Gvk: resid.Gvk{
				Kind: "ConfigMap",
			},
			Referrers: fsSlice1,
		},
		{
			Gvk: resid.Gvk{
				Kind: "Secret",
			},
			Referrers: fsSlice2,
		},
	}
	expected := nbrSlice{
		{
			Gvk: resid.Gvk{
				Kind: "ConfigMap",
			},
			Referrers: fsSlice1,
		},
		{
			Gvk: resid.Gvk{
				Kind: "Secret",
			},
			Referrers: fsSlice2,
		},
	}
	actual, err := nbrsSlice1.mergeAll(nbrsSlice2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected\n %v\n but got\n %v\n", expected, actual)
	}
}

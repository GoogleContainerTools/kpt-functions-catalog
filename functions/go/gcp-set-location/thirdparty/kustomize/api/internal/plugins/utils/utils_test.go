// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestDeterminePluginSrcRoot(t *testing.T) {
	actual, err := DeterminePluginSrcRoot(filesys.MakeFsOnDisk())
	if err != nil {
		t.Error(err)
	}
	if !filepath.IsAbs(actual) {
		t.Errorf("expected absolute path, but got '%s'", actual)
	}
	if !strings.HasSuffix(actual, konfig.RelPluginHome) {
		t.Errorf("expected suffix '%s' in '%s'", konfig.RelPluginHome, actual)
	}
}

func makeConfigMap(rf *resource.Factory, name, behavior string, hashValue *string) *resource.Resource {
	r := rf.FromMap(map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata":   map[string]interface{}{"name": name},
	})
	annotations := map[string]string{}
	if behavior != "" {
		annotations[BehaviorAnnotation] = behavior
	}
	if hashValue != nil {
		annotations[HashAnnotation] = *hashValue
	}
	if len(annotations) > 0 {
		if err := r.SetAnnotations(annotations); err != nil {
			panic(err)
		}
	}
	return r
}

func makeConfigMapOptions(rf *resource.Factory, name, behavior string, disableHash bool) *resource.Resource {
	return rf.FromMapAndOption(map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata":   map[string]interface{}{"name": name},
	}, &types.GeneratorArgs{
		Behavior: behavior,
		Options:  &types.GeneratorOptions{DisableNameSuffixHash: disableHash}})
}

func strptr(s string) *string {
	return &s
}

func TestUpdateResourceOptions(t *testing.T) {
	rf := provider.NewDefaultDepProvider().GetResourceFactory()
	in := resmap.New()
	expected := resmap.New()
	cases := []struct {
		behavior  string
		needsHash bool
		hashValue *string
	}{
		{hashValue: strptr("false")},
		{hashValue: strptr("true"), needsHash: true},
		{behavior: "replace"},
		{behavior: "merge"},
		{behavior: "create"},
		{behavior: "nonsense"},
		{behavior: "merge", hashValue: strptr("false")},
		{behavior: "merge", hashValue: strptr("true"), needsHash: true},
	}
	for i, c := range cases {
		name := fmt.Sprintf("test%d", i)
		err := in.Append(makeConfigMap(rf, name, c.behavior, c.hashValue))
		require.NoError(t, err)
		err = expected.Append(makeConfigMapOptions(rf, name, c.behavior, !c.needsHash))
		require.NoError(t, err)
	}
	actual, err := UpdateResourceOptions(in)
	assert.NoError(t, err)
	assert.NoError(t, expected.ErrorIfNotEqualLists(actual))
}

func TestUpdateResourceOptionsWithInvalidHashAnnotationValues(t *testing.T) {
	rf := provider.NewDefaultDepProvider().GetResourceFactory()
	cases := []string{
		"",
		"FaLsE",
		"TrUe",
		"potato",
	}
	for i, c := range cases {
		name := fmt.Sprintf("test%d", i)
		in := resmap.New()
		err := in.Append(makeConfigMap(rf, name, "", &c))
		require.NoError(t, err)
		_, err = UpdateResourceOptions(in)
		require.Error(t, err)
	}
}

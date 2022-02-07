// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	"sigs.k8s.io/kustomize/api/types"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// Demo custom configuration of a builtin transformation.
// This is a NamePrefixer that touches Deployments
// and Services exclusively.
func TestCustomNamePrefixer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	th.GetPluginConfig().BpLoadingOptions = types.BploUseStaticallyLinked
	defer th.Reset()

	th.WriteK(".", `
resources:
- deployment.yaml
- role.yaml
- service.yaml
transformers:
- prefixer.yaml
`)
	th.WriteF("prefixer.yaml", `
apiVersion: builtin
kind: PrefixSuffixTransformer
metadata:
  name: customPrefixer
prefix: zzz-
fieldSpecs:
- kind: Deployment
  path: metadata/name
- kind: Service
  path: metadata/name
`)
	th.WriteF("deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    metadata:
      labels:
        backend: awesome
    spec:
      containers:
      - name: whatever
        image: whatever
`)
	th.WriteF("role.yaml", `
apiVersion: v1
kind: Role
metadata:
  name: myRole
`)
	th.WriteF("service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService
`)

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: zzz-myDeployment
spec:
  template:
    metadata:
      labels:
        backend: awesome
    spec:
      containers:
      - image: whatever
        name: whatever
---
apiVersion: v1
kind: Role
metadata:
  name: myRole
---
apiVersion: v1
kind: Service
metadata:
  name: zzz-myService
`)
}

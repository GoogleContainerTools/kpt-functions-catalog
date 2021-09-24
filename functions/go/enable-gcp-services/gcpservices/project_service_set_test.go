package gcpservices

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
)

func TestProjectServiceSet_Filter(t *testing.T) {
	var tests = []struct {
		name        string
		resourceMap map[string]string
		expected    string
		results     []framework.ResultItem
		errMsg      string
	}{
		{
			name: "simple",
			resourceMap: map[string]string{"ps.yaml": `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services
  annotations:
    config.kubernetes.io/local-config: "true"
spec:
  services:
  - compute.googleapis.com
  projectID: test
`},
			expected: `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services
  annotations:
    config.kubernetes.io/local-config: "true"
    config.kubernetes.io/path: 'ps.yaml'
spec:
  services:
  - compute.googleapis.com
  projectID: test
---
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-compute
  annotations:
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/project-services'
    config.kubernetes.io/path: 'service_project-services-compute.yaml'
spec:
  resourceID: compute.googleapis.com
  projectRef:
    external: test
`,
			results: []framework.ResultItem{getResult(generateAction, "project-services-compute", "", "service_project-services-compute.yaml")},
		},
		{
			name: "simple no project, without local config anno",
			resourceMap: map[string]string{"ps.yaml": `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services
spec:
  services:
  - compute.googleapis.com
  - redis.googleapis.com
`},
			expected: `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services
  annotations:
    config.kubernetes.io/path: 'ps.yaml'
    config.kubernetes.io/local-config: 'true'
spec:
  services:
  - compute.googleapis.com
  - redis.googleapis.com
---
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-compute
  annotations:
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/project-services'
    config.kubernetes.io/path: 'service_project-services-compute.yaml'
spec:
  resourceID: compute.googleapis.com
---
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-redis
  annotations:
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/project-services'
    config.kubernetes.io/path: 'service_project-services-redis.yaml'
spec:
  resourceID: redis.googleapis.com
`,
			results: []framework.ResultItem{
				getResult(generateAction, "project-services-compute", "", "service_project-services-compute.yaml"),
				getResult(generateAction, "project-services-redis", "", "service_project-services-redis.yaml"),
			},
		},
		{
			name: "simple with annotations",
			resourceMap: map[string]string{"ps.yaml": `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: "false"
    config.kubernetes.io/local-config: "true"
spec:
  services:
  - compute.googleapis.com
  projectID: test
`},
			expected: `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: "false"
    config.kubernetes.io/local-config: "true"
    config.kubernetes.io/path: 'ps.yaml'
spec:
  services:
  - compute.googleapis.com
  projectID: test
---
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-compute
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: 'false'
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/project-services'
    config.kubernetes.io/path: 'service_project-services-compute.yaml'
spec:
  resourceID: compute.googleapis.com
  projectRef:
    external: test
`,
			results: []framework.ResultItem{getResult(generateAction, "project-services-compute", "", "service_project-services-compute.yaml")},
		},
		{
			name: "simple with annotations with ns",
			resourceMap: map[string]string{"ps.yaml": `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services
  namespace: foo
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: "false"
    config.kubernetes.io/local-config: "true"
spec:
  services:
  - compute.googleapis.com
  projectID: test
`},
			expected: `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services
  namespace: foo
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: "false"
    config.kubernetes.io/local-config: "true"
    config.kubernetes.io/path: 'ps.yaml'
spec:
  services:
  - compute.googleapis.com
  projectID: test
---
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-compute
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: 'false'
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/project-services'
    config.kubernetes.io/path: 'foo/service_project-services-compute.yaml'
  namespace: foo
spec:
  resourceID: compute.googleapis.com
  projectRef:
    external: test
`,
			results: []framework.ResultItem{getResult(generateAction, "project-services-compute", "foo", "foo/service_project-services-compute.yaml")},
		},
		{
			name: "simple with existing service generated",
			resourceMap: map[string]string{"ps.yaml": `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services
  annotations:
    config.kubernetes.io/local-config: "true"
    new: anno
spec:
  services:
  - compute.googleapis.com
  projectID: test
`, "compute.yaml": `apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-compute
  annotations:
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/project-services'
spec:
  resourceID: compute.googleapis.com
  projectRef:
    external: test`},
			expected: `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services
  annotations:
    config.kubernetes.io/local-config: "true"
    new: anno
    config.kubernetes.io/path: 'ps.yaml'
spec:
  services:
  - compute.googleapis.com
  projectID: test
---
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-compute
  annotations:
    new: 'anno'
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/project-services'
    config.kubernetes.io/path: 'service_project-services-compute.yaml'
spec:
  resourceID: compute.googleapis.com
  projectRef:
    external: test
`,
			results: []framework.ResultItem{getResult(recreateAction, "project-services-compute", "", "service_project-services-compute.yaml")},
		},

		{
			name: "simple with new service, other objects and pruning previously generated services",
			resourceMap: map[string]string{
				"ps.yaml": `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services
  annotations:
    config.kubernetes.io/local-config: "true"
spec:
  services:
  - redis.googleapis.com
  projectID: test`,
				"bq.yaml": `apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-bigquery
  annotations:
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/project-services'
spec:
  resourceID: bigquery.googleapis.com
  projectRef:
    external: test`,
				"compute.yaml": `apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-compute
  namespace: foo
  annotations:
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/project-services'
spec:
  resourceID: compute.googleapis.com
  projectRef:
    external: test`,
				"deploy1.yaml": `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app1
  name: mungebot1`},
			expected: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app1
  name: mungebot1
  annotations:
    config.kubernetes.io/path: 'deploy1.yaml'
---
apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services
  annotations:
    config.kubernetes.io/local-config: "true"
    config.kubernetes.io/path: 'ps.yaml'
spec:
  services:
  - redis.googleapis.com
  projectID: test
---
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-redis
  annotations:
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/project-services'
    config.kubernetes.io/path: 'service_project-services-redis.yaml'
spec:
  resourceID: redis.googleapis.com
  projectRef:
    external: test
`,
			results: []framework.ResultItem{
				getResult(generateAction, "project-services-redis", "", "service_project-services-redis.yaml"),
				getResult(pruneAction, "project-services-compute", "foo", "compute.yaml"),
				getResult(pruneAction, "project-services-bigquery", "", "bq.yaml"),
			},
		},
		{
			name: "multiple with annotations with ns",
			resourceMap: map[string]string{"ps1.yaml": `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services-one
  namespace: foo
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: "false"
    config.kubernetes.io/local-config: "true"
spec:
  services:
  - compute.googleapis.com
  projectID: test
`, "ps2.yaml": `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services-two
  namespace: bar
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: "false"
    config.kubernetes.io/local-config: "true"
spec:
  services:
  - redis.googleapis.com
  projectID: test
`},
			expected: `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services-one
  namespace: foo
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: "false"
    config.kubernetes.io/local-config: "true"
    config.kubernetes.io/path: 'ps1.yaml'
spec:
  services:
  - compute.googleapis.com
  projectID: test
---
apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services-two
  namespace: bar
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: "false"
    config.kubernetes.io/local-config: "true"
    config.kubernetes.io/path: 'ps2.yaml'
spec:
  services:
  - redis.googleapis.com
  projectID: test
---
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-one-compute
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: 'false'
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/project-services-one'
    config.kubernetes.io/path: 'foo/service_project-services-one-compute.yaml'
  namespace: foo
spec:
  resourceID: compute.googleapis.com
  projectRef:
    external: test
---
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-two-redis
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: 'false'
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/project-services-two'
    config.kubernetes.io/path: 'bar/service_project-services-two-redis.yaml'
  namespace: bar
spec:
  resourceID: redis.googleapis.com
  projectRef:
    external: test
`,
			results: []framework.ResultItem{
				getResult(generateAction, "project-services-one-compute", "foo", "foo/service_project-services-one-compute.yaml"),
				getResult(generateAction, "project-services-two-redis", "bar", "bar/service_project-services-two-redis.yaml"),
			},
		},
		{
			name: "multiple with prune an existing service",
			resourceMap: map[string]string{"ps1.yaml": `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services-one
  namespace: foo
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: "false"
    config.kubernetes.io/local-config: "true"
spec:
  services:
  - compute.googleapis.com
  projectID: test
`, "ps2.yaml": `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services-two
  namespace: bar
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: "false"
    config.kubernetes.io/local-config: "true"
spec:
  services:
  - redis.googleapis.com
  projectID: test
`, "bar/service_project-services-two-redis.yaml": `---
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-two-redis
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: 'false'
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/project-services-two'
  namespace: bar
spec:
  resourceID: redis.googleapis.com
  projectRef:
    external: test
`},
			expected: `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services-one
  namespace: foo
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: "false"
    config.kubernetes.io/local-config: "true"
    config.kubernetes.io/path: 'ps1.yaml'
spec:
  services:
  - compute.googleapis.com
  projectID: test
---
apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services-two
  namespace: bar
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: "false"
    config.kubernetes.io/local-config: "true"
    config.kubernetes.io/path: 'ps2.yaml'
spec:
  services:
  - redis.googleapis.com
  projectID: test
---
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-one-compute
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: 'false'
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/project-services-one'
    config.kubernetes.io/path: 'foo/service_project-services-one-compute.yaml'
  namespace: foo
spec:
  resourceID: compute.googleapis.com
  projectRef:
    external: test
---
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-two-redis
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: 'false'
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/project-services-two'
    config.kubernetes.io/path: 'bar/service_project-services-two-redis.yaml'
  namespace: bar
spec:
  resourceID: redis.googleapis.com
  projectRef:
    external: test
`,
			results: []framework.ResultItem{
				getResult(generateAction, "project-services-one-compute", "foo", "foo/service_project-services-one-compute.yaml"),
				getResult(recreateAction, "project-services-two-redis", "bar", "bar/service_project-services-two-redis.yaml"),
			},
		},
		{
			name: "multiple in different packages",
			resourceMap: map[string]string{"ps1.yaml": `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services-one
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: "false"
    config.kubernetes.io/local-config: "true"
spec:
  services:
  - compute.googleapis.com
  projectID: test
`, "subpkg/ps2.yaml": `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services-two
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: "false"
    config.kubernetes.io/local-config: "true"
spec:
  services:
  - redis.googleapis.com
  projectID: test
`},
			expected: `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services-one
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: "false"
    config.kubernetes.io/local-config: "true"
    config.kubernetes.io/path: 'ps1.yaml'
spec:
  services:
  - compute.googleapis.com
  projectID: test
---
apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services-two
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: "false"
    config.kubernetes.io/local-config: "true"
    config.kubernetes.io/path: 'subpkg/ps2.yaml'
spec:
  services:
  - redis.googleapis.com
  projectID: test
---
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-one-compute
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: 'false'
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/project-services-one'
    config.kubernetes.io/path: 'service_project-services-one-compute.yaml'
spec:
  resourceID: compute.googleapis.com
  projectRef:
    external: test
---
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-two-redis
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: 'false'
    blueprints.cloud.google.com/ownerReference: 'blueprints.cloud.google.com/ProjectServiceSet/project-services-two'
    config.kubernetes.io/path: 'subpkg/service_project-services-two-redis.yaml'
spec:
  resourceID: redis.googleapis.com
  projectRef:
    external: test
`,
			results: []framework.ResultItem{
				getResult(generateAction, "project-services-one-compute", "", "service_project-services-one-compute.yaml"),
				getResult(generateAction, "project-services-two-redis", "", "subpkg/service_project-services-two-redis.yaml"),
			},
		},
		{
			name: "invalid empty",
			resourceMap: map[string]string{"ps.yaml": `apiVersion: blueprints.cloud.google.com/v1alpha1
kind: ProjectServiceSet
metadata:
  name: project-services
spec:
  services: []
  projectID: test
`},
			errMsg: "at least one service must be specified under `spec.services[]`",
		},
		{
			name: "no project services CR noop",
			resourceMap: map[string]string{"compute.yaml": `apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: custom-compute
spec:
  resourceID: compute.googleapis.com
  projectRef:
    external: test`, "deploy1.yaml": `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app1
  name: mungebot1`},
			expected: `apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: custom-compute
  annotations:
    config.kubernetes.io/path: 'compute.yaml'
spec:
  resourceID: compute.googleapis.com
  projectRef:
    external: test
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app1
  name: mungebot1
  annotations:
    config.kubernetes.io/path: 'deploy1.yaml'
`,
			results: []framework.ResultItem{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			pkgDir := setupInputs(t, tt.resourceMap)
			defer os.RemoveAll(pkgDir)
			pslr := ProjectServiceSetRunner{}
			in := &kio.LocalPackageReader{
				PackagePath: pkgDir,
			}
			out := &bytes.Buffer{}
			err := kio.Pipeline{
				Inputs:  []kio.Reader{in},
				Filters: []kio.Filter{&pslr},
				Outputs: []kio.Writer{kio.ByteWriter{Writer: out}},
			}.Execute()

			if tt.errMsg != "" {
				require.NotNil(err)
				require.Contains(err.Error(), tt.errMsg)
			} else {
				require.NoError(err)
				require.Equal(tt.expected, out.String())
				require.ElementsMatch(tt.results, pslr.GetResults())
			}

		})
	}
}

func setupInputs(t *testing.T, resourceMap map[string]string) string {
	t.Helper()
	require := require.New(t)
	baseDir, err := ioutil.TempDir("", "")
	require.NoError(err)

	for rpath, data := range resourceMap {
		filePath := path.Join(baseDir, rpath)
		err = os.MkdirAll(path.Dir(filePath), os.ModePerm)
		require.NoError(err)
		err = ioutil.WriteFile(path.Join(baseDir, rpath), []byte(data), 0644)
		require.NoError(err)
	}
	return baseDir
}

func getResult(action actionType, name, ns, fp string) framework.ResultItem {
	r := framework.ResultItem{
		File:     framework.File{Path: fp},
		Message:  action.String(),
		Severity: framework.Info}
	r.ResourceRef.Name = name
	r.ResourceRef.Namespace = ns
	r.ResourceRef.APIVersion = serviceUsageAPIVersion
	r.ResourceRef.Kind = serviceUsageKind
	return r
}

func TestProjectServiceSet_validate(t *testing.T) {
	tests := []struct {
		name        string
		apiVersion  string
		kind        string
		services    []string
		annotations map[string]string
		errMsg      string
	}{
		{
			name:        "valid",
			apiVersion:  projectServiceSetAPIVersion,
			kind:        projectServiceSetKind,
			annotations: map[string]string{filters.LocalConfigAnnotation: "true"},
			services:    []string{"compute.googleapis.com"},
		},
		{
			name:        "empty services",
			apiVersion:  projectServiceSetAPIVersion,
			kind:        projectServiceSetKind,
			services:    []string{},
			annotations: map[string]string{filters.LocalConfigAnnotation: "true"},
			errMsg:      "at least one service must be specified under `spec.services[]`",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			psl := ProjectServiceSet{}
			psl.APIVersion = tt.apiVersion
			psl.Kind = tt.kind
			psl.Spec.Services = tt.services
			psl.Annotations = tt.annotations
			err := psl.validate()
			if tt.errMsg != "" {
				require.NotNil(err)
				require.Equal(tt.errMsg, err.Error())
			} else {
				require.NoError(err)
			}
		})
	}
}

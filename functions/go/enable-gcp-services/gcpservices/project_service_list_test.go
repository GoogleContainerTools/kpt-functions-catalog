package gcpservices

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func TestProjectServiceList_Filter(t *testing.T) {
	var tests = []struct {
		name        string
		fnConfig    ProjectServiceList
		resourceMap map[string]string
		expected    string
		results     []Result
		errMsg      string
	}{
		{
			name:     "simple",
			fnConfig: getProjectServiceList("project-services", []string{"compute.googleapis.com"}, "test", "", nil),
			expected: `apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-compute
  annotations:
    blueprints.cloud.google.com/managed-by-enable-gcp-services: 'project-services'
spec:
  resourceID: compute.googleapis.com
  projectRef:
    external: test
`,
			results: []Result{getResult("generated service", "project-services-compute", "", "")},
		},
		{
			name:     "simple no project",
			fnConfig: getProjectServiceList("project-services", []string{"compute.googleapis.com", "redis.googleapis.com"}, "", "", nil),
			expected: `apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-compute
  annotations:
    blueprints.cloud.google.com/managed-by-enable-gcp-services: 'project-services'
spec:
  resourceID: compute.googleapis.com
---
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-redis
  annotations:
    blueprints.cloud.google.com/managed-by-enable-gcp-services: 'project-services'
spec:
  resourceID: redis.googleapis.com
`,
			results: []Result{
				getResult("generated service", "project-services-compute", "", ""),
				getResult("generated service", "project-services-redis", "", ""),
			},
		},
		{
			name:     "simple with annotations",
			fnConfig: getProjectServiceList("project-services", []string{"compute.googleapis.com"}, "test", "", map[string]string{"cnrm.cloud.google.com/disable-dependent-services": "false"}),
			expected: `apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-compute
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: 'false'
    blueprints.cloud.google.com/managed-by-enable-gcp-services: 'project-services'
spec:
  resourceID: compute.googleapis.com
  projectRef:
    external: test
`,
			results: []Result{getResult("generated service", "project-services-compute", "", "")},
		},
		{
			name:     "simple with annotations with ns",
			fnConfig: getProjectServiceList("project-services", []string{"compute.googleapis.com"}, "test", "foo", map[string]string{"cnrm.cloud.google.com/disable-dependent-services": "false"}),
			expected: `apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-compute
  annotations:
    cnrm.cloud.google.com/disable-dependent-services: 'false'
    blueprints.cloud.google.com/managed-by-enable-gcp-services: 'project-services'
  namespace: foo
spec:
  resourceID: compute.googleapis.com
  projectRef:
    external: test
`,
			results: []Result{getResult("generated service", "project-services-compute", "foo", "")},
		},
		{
			name:     "simple with existing service generated",
			fnConfig: getProjectServiceList("project-services", []string{"compute.googleapis.com"}, "test", "", map[string]string{"new": "anno"}),
			expected: `apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-compute
  annotations:
    new: 'anno'
    blueprints.cloud.google.com/managed-by-enable-gcp-services: 'project-services'
spec:
  resourceID: compute.googleapis.com
  projectRef:
    external: test
`, resourceMap: map[string]string{"compute.yaml": `apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-compute
  annotations:
    blueprints.cloud.google.com/managed-by-enable-gcp-services: 'project-services'
spec:
  resourceID: compute.googleapis.com
  projectRef:
    external: test`},
			results: []Result{
				getResult("generated service", "project-services-compute", "", ""),
				getResult("pruned service", "project-services-compute", "", "compute.yaml"),
			},
		},

		{
			name:     "simple with new service, other objects and pruning previously generated services",
			fnConfig: getProjectServiceList("project-services", []string{"redis.googleapis.com"}, "test", "", nil),
			expected: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app1
  name: mungebot1
  annotations:
    config.kubernetes.io/path: 'deploy1.yaml'
---
apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-redis
  annotations:
    blueprints.cloud.google.com/managed-by-enable-gcp-services: 'project-services'
spec:
  resourceID: redis.googleapis.com
  projectRef:
    external: test
`, resourceMap: map[string]string{
				"bq.yaml": `apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: project-services-bigquery
  annotations:
    blueprints.cloud.google.com/managed-by-enable-gcp-services: 'project-services'
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
    blueprints.cloud.google.com/managed-by-enable-gcp-services: 'project-services'
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
			results: []Result{
				getResult("generated service", "project-services-redis", "", ""),
				getResult("pruned service", "project-services-compute", "foo", "compute.yaml"),
				getResult("pruned service", "project-services-bigquery", "", "bq.yaml"),
			},
		},
		{
			name:     "invalid empty",
			fnConfig: getProjectServiceList("project-services", []string{}, "test", "", nil),
			errMsg:   "at least one service must be specified under spec.services[]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			pkgDir := setupInputs(t, tt.resourceMap)
			defer os.RemoveAll(pkgDir)

			in := &kio.LocalPackageReader{
				PackagePath: pkgDir,
			}
			out := &bytes.Buffer{}
			err := kio.Pipeline{
				Inputs:  []kio.Reader{in},
				Filters: []kio.Filter{&tt.fnConfig},
				Outputs: []kio.Writer{kio.ByteWriter{Writer: out}},
			}.Execute()

			if tt.errMsg != "" {
				require.NotNil(err)
				require.Contains(err.Error(), tt.errMsg)
			} else {
				require.NoError(err)
				require.Equal(tt.expected, out.String())
				require.ElementsMatch(tt.results, tt.fnConfig.GetResults())
			}

		})
	}
}

func getProjectServiceList(name string, services []string, projectID string, ns string, annotations map[string]string) ProjectServiceList {
	p := ProjectServiceList{}
	p.APIVersion = projectServiceListAPIVersion
	p.Kind = projectServiceListKind
	p.Name = name
	p.Spec.Services = services
	p.Spec.ProjectID = projectID
	p.Namespace = ns
	p.Annotations = annotations
	return p
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

func getResult(action, name, ns, fp string) Result {
	r := Result{Action: action, FilePath: fp}
	r.ResourceRef.Name = name
	r.ResourceRef.Namespace = ns
	r.ResourceRef.APIVersion = serviceUsageAPIVersion
	r.ResourceRef.Kind = serviceUsageKind
	return r
}

func TestProjectServiceList_validate(t *testing.T) {
	tests := []struct {
		name       string
		apiVersion string
		kind       string
		services   []string
		errMsg     string
	}{
		{
			name:       "valid",
			apiVersion: projectServiceListAPIVersion,
			kind:       projectServiceListKind,
			services:   []string{"compute.googleapis.com"},
		},
		{
			name:       "invalid api version",
			apiVersion: "foo",
			errMsg:     "invalid APIVersion: foo supported APIVersion: blueprints.cloud.google.com/v1alpha1",
		},
		{
			name:       "invalid kind",
			apiVersion: projectServiceListAPIVersion,
			kind:       "foo",
			errMsg:     "invalid Kind: foo supported Kind: ProjectServiceList",
		},
		{
			name:       "empty services",
			apiVersion: projectServiceListAPIVersion,
			kind:       projectServiceListKind,
			services:   []string{},
			errMsg:     "at least one service must be specified under spec.services[]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			psl := ProjectServiceList{}
			psl.APIVersion = tt.apiVersion
			psl.Kind = tt.kind
			psl.Spec.Services = tt.services
			err := psl.validate()
			if tt.errMsg != "" {
				require.NotNil(err)
				require.Contains(err.Error(), tt.errMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

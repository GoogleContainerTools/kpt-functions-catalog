package listsetters

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func TestListSettersFilter(t *testing.T) {
	var tests = []struct {
		name           string
		input          string
		kf             string
		setterYml      string
		expectedResult []*Result
		errMsg         string
		warnings       []*ErrSetterDiscovery
	}{
		{
			name: "No setters",
			input: `apiVersion: v1
kind: Service
metadata:
  name: my-app
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app
  name: mungebot`,
			expectedResult: []*Result{},
			warnings:       []*ErrSetterDiscovery{{"unable to find Kptfile, please include --include-meta-resources flag if a Kptfile is present"}},
		},
		{
			name: "Scalar Simple",
			kf: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configMap:
        app: my-app`,
			input: `apiVersion: v1
kind: Service
metadata:
  name: my-app # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app # kpt-set: ${app}
  name: mungebot`,
			expectedResult: []*Result{{Name: "app", Value: "my-app", Count: 2, Type: "string"}},
		},
		{
			name: "Scalar Simple invalid kf",
			kf: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test
foo: bar`,
			input: `apiVersion: v1
kind: Service
metadata:
  name: my-app # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app # kpt-set: ${app}
  name: mungebot`,
			errMsg: "unable to read Kptfile: please make sure the package has a valid 'v1' Kptfile: yaml: unmarshal errors:\n  line 8: field foo not found in type v1.KptFile",
		},
		{
			name: "Scalar Simple missing apply-setters",
			kf: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/foo:v0.1
      configMap:
        app: my-app`,
			input: `apiVersion: v1
kind: Service
metadata:
  name: my-app # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app # kpt-set: ${app}
  name: mungebot`,
			expectedResult: []*Result{{Name: "app", Value: "my-app", Count: 2, Type: "string"}},
			warnings:       []*ErrSetterDiscovery{{"unable to find apply-setters fn in Kptfile Pipeline.Mutators"}},
		},
		{
			name: "Scalar Simple missing kf pipeline",
			kf: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test`,
			input: `apiVersion: v1
kind: Service
metadata:
  name: my-app # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app # kpt-set: ${app}
  name: mungebot`,
			expectedResult: []*Result{{Name: "app", Value: "my-app", Count: 2, Type: "string"}},
			warnings:       []*ErrSetterDiscovery{{"unable to find Pipeline declaration in Kptfile"}},
		},
		{
			name: "Scalar Simple no apply-setter fnConfig",
			kf: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1`,
			input: `apiVersion: v1
kind: Service
metadata:
  name: my-app # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app # kpt-set: ${app}
  name: mungebot`,
			expectedResult: []*Result{{Name: "app", Value: "my-app", Count: 2, Type: "string"}},
			warnings:       []*ErrSetterDiscovery{{"unable to find ConfigMap or ConfigPath fnConfig for apply-setters"}},
		},
		{
			name: "Scalar Simple missing apply-setter configPath file",
			kf: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configPath: setters.yaml`,
			input: `apiVersion: v1
kind: Service
metadata:
  name: my-app # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app # kpt-set: ${app}
  name: mungebot`,
			expectedResult: []*Result{{Name: "app", Value: "my-app", Count: 2, Type: "string"}},
			errMsg:         "file setters.yaml doesn't exist, please ensure the file specified in \"configPath\" exists and retry",
		},
		{
			name: "Scalar with zero count setter",
			kf: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configMap:
        app: my-app
        foo: bar
`,
			input: `apiVersion: v1
kind: Service
metadata:
  name: my-app # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app # kpt-set: ${app}
  name: mungebot			
`,
			expectedResult: []*Result{{Name: "app", Value: "my-app", Count: 2, Type: "string"}, {Name: "foo", Value: "bar", Count: 0, Type: "string"}},
		},
		{
			name: "Mapping Simple",
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: # kpt-set: ${images}
    - ubuntu
    - hbase
 `,
			expectedResult: []*Result{{Name: "images", Value: "[hbase ubuntu]", Count: 1, Type: "list"}},
			warnings:       []*ErrSetterDiscovery{{"unable to find Kptfile, please include --include-meta-resources flag if a Kptfile is present"}},
		},
		{
			name: "Mapping with kptfile and setterYml",
			kf: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configPath: setters.yaml`,
			setterYml: `apiVersion: v1
kind: ConfigMap
metadata:
  name: apply-setters-fn-config
data:
  images: |
      - ubuntu
      - hbase
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: # kpt-set: ${images}
    - ubuntu
    - hbase
 `,
			expectedResult: []*Result{{Name: "images", Value: "[ubuntu hbase]", Count: 1, Type: "list"}},
		},
		{
			name: "Scalar and Mapping",
			input: `apiVersion: dns.cnrm.cloud.google.com/v1beta1
kind: DNSRecordSet
metadata:
  name: dnsrecordset-sample-mx # kpt-set: ${record-set-name}
  labels:
    typettl: "MX-300" # kpt-set: ${type}-${ttl}
spec:
  name: "mail.example.com." # kpt-set: ${domain}
  type: "MX" # kpt-set: ${type}
  ttl: 300 # kpt-set: ${ttl}
  managedZoneRef:
    name: dnsrecordset-dep-mx # kpt-set: ${managed-zone-name}
  rrdatas: # kpt-set: ${records}
    - "5 gmr-stmp-in.l.google.com."
    - "10 alt1.gmr-stmp-in.l.google.com."
    - "10 alt2.gmr-stmp-in.l.google.com."
    - "10 alt3.gmr-stmp-in.l.google.com."
    - "10 alt4.gmr-stmp-in.l.google.com."
`,
			expectedResult: []*Result{
				{Name: "record-set-name", Value: "dnsrecordset-sample-mx", Count: 1, Type: "string"},
				{Name: "type", Value: "MX", Count: 2, Type: "string"},
				{Name: "domain", Value: "mail.example.com.", Count: 1, Type: "string"},
				{Name: "managed-zone-name", Value: "dnsrecordset-dep-mx", Count: 1, Type: "string"},
				{Name: "ttl", Value: "300", Count: 2, Type: "string"},
				{Name: "records", Value: "[10 alt1.gmr-stmp-in.l.google.com. 10 alt2.gmr-stmp-in.l.google.com. 10 alt3.gmr-stmp-in.l.google.com. 10 alt4.gmr-stmp-in.l.google.com. 5 gmr-stmp-in.l.google.com.]", Count: 1, Type: "list"},
			},
			warnings: []*ErrSetterDiscovery{{"unable to find Kptfile, please include --include-meta-resources flag if a Kptfile is present"}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require := require.New(t)
			pkgDir := setupInputs(t, test.input, test.kf, test.setterYml)
			defer os.RemoveAll(pkgDir)

			ls := New()
			inout := &kio.LocalPackageReadWriter{
				PackagePath:     pkgDir,
				NoDeleteFiles:   true,
				PackageFileName: "Kptfile",
				MatchFilesGlob:  append(kio.DefaultMatch, "Kptfile"),
			}
			err := kio.Pipeline{
				Inputs:  []kio.Reader{inout},
				Filters: []kio.Filter{&ls},
				Outputs: []kio.Writer{inout},
			}.Execute()
			if test.errMsg != "" {
				require.NotNil(err)
				require.Contains(err.Error(), test.errMsg)
			} else {
				require.NoError(err)
				require.ElementsMatch(ls.Warnings, test.warnings)
				actualResources := ls.GetResults()
				require.ElementsMatch(actualResources, test.expectedResult)
			}

		})
	}
}

func setupInputs(t *testing.T, input, kf, setterYml string) string {
	t.Helper()
	require := require.New(t)
	baseDir, err := ioutil.TempDir("", "")
	require.NoError(err)

	r, err := ioutil.TempFile(baseDir, "k8s-cli-*.yaml")
	require.NoError(err)
	err = ioutil.WriteFile(r.Name(), []byte(input), 0600)

	if kf != "" {
		err = ioutil.WriteFile(path.Join(baseDir, "Kptfile"), []byte(kf), 0644)
		require.NoError(err)
	}

	if setterYml != "" {
		err = ioutil.WriteFile(path.Join(baseDir, "setters.yaml"), []byte(setterYml), 0644)
		require.NoError(err)
	}
	require.NoError(err)
	return baseDir
}

func TestCurrentSetterValues(t *testing.T) {
	var tests = []struct {
		name     string
		value    string
		pattern  string
		expected map[string]string
	}{
		{
			name:    "setter values from pattern 1",
			value:   "foo-dev-bar-us-east-1-baz",
			pattern: `foo-${environment}-bar-${region}-baz`,
			expected: map[string]string{
				"environment": "dev",
				"region":      "us-east-1",
			},
		},
		{
			name:    "setter values from pattern 2",
			value:   "foo-dev-bar-us-east-1-baz",
			pattern: `foo-${environment}-bar-${region}-baz`,
			expected: map[string]string{
				"environment": "dev",
				"region":      "us-east-1",
			},
		},
		{
			name:    "setter values from pattern 3",
			value:   "gcr.io/my-app/my-app-backend:1.0.0",
			pattern: `${registry}/${app~!@#$%^&*()<>?:"|}/${app-image-name}:${app-image-tag}`,
			expected: map[string]string{
				"registry":             "gcr.io",
				`app~!@#$%^&*()<>?:"|`: "my-app",
				"app-image-name":       "my-app-backend",
				"app-image-tag":        "1.0.0",
			},
		},
		{
			name:     "setter values from pattern unresolved",
			value:    "foo-dev-bar-us-east-1-baz",
			pattern:  `${image}:${tag}`,
			expected: map[string]string{},
		},
		{
			name:     "setter values from pattern unresolved 2",
			value:    "nginx:1.2",
			pattern:  `${image}${tag}`,
			expected: map[string]string{},
		},
		{
			name:     "setter values from pattern unresolved 3",
			value:    "my-project/nginx:1.2",
			pattern:  `${project-id}/${image}${tag}`,
			expected: map[string]string{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := currentSetterValues(test.pattern, test.value)
			require.Equal(t, test.expected, res)
		})
	}
}

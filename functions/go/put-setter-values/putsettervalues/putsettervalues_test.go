package putsettervalues

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestPutSetterValuesFilter(t *testing.T) {
	var tests = []struct {
		name                  string
		config                string
		kptfile               string
		settersConfig         string
		expectedKptfile       string
		expectedSettersConfig string
		errMsg                string
	}{
		{
			name: "put values in setters.yaml",
			kptfile: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: nginx
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configPath: setters.yaml
`,
			settersConfig: `apiVersion: v1
kind: ConfigMap
metadata:
  name: setters
data:
  namespace: some-space
  image: nginx
  list: |
    - dev
    - stage
`,
			config: `
data:
  namespace: my-space
  image: ubuntu
  list: |
    - dev
    - prod
  tag: 1.14.2
`,
			expectedKptfile: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: nginx
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configPath: setters.yaml
`,
			expectedSettersConfig: `apiVersion: v1
kind: ConfigMap
metadata:
  name: setters
data:
  image: ubuntu
  list: |
    - dev
    - prod
  namespace: my-space
  tag: 1.14.2
`,
		},
		{
			name: "put values in Kptfile",
			kptfile: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: nginx
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configMap:
        image: nginx
        namespace: some-space
`,
			config: `
data:
  namespace: my-space
  image: ubuntu
  list: |
    - dev
    - prod
  tag: 1.14.2
`,
			expectedKptfile: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: nginx
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configMap:
        image: ubuntu
        list: |
          - dev
          - prod
        namespace: my-space
        tag: 1.14.2
`,
		},
		{
			name: "put values with multiple declarations, put values in all declarations",
			kptfile: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: nginx
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configPath: setters.yaml
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configMap:
        image: nginx
        namespace: some-space
`,
			settersConfig: `apiVersion: v1
kind: ConfigMap
metadata:
  name: setters
data:
  namespace: some-space
  image: nginx
  list: |
    - dev
    - stage
`,
			config: `
data:
  namespace: my-space
  image: ubuntu
  list: |
    - dev
    - prod
  tag: 1.14.2
`,
			expectedKptfile: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: nginx
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configPath: setters.yaml
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configMap:
        image: ubuntu
        list: |
          - dev
          - prod
        namespace: my-space
        tag: 1.14.2
`,
			expectedSettersConfig: `apiVersion: v1
kind: ConfigMap
metadata:
  name: setters
data:
  image: ubuntu
  list: |
    - dev
    - prod
  namespace: my-space
  tag: 1.14.2
`,
		},
		{
			name: "no Kptfile",
			config: `
data:
  namespace: my-space
  image: ubuntu
  list: |
    - dev
    - prod
  tag: 1.14.2
`,
			errMsg: `unable to find "Kptfile" in the package, please ensure "Kptfile" is present in the root directory and specify --include-meta-resources flag`,
		},
		{
			name: "no setters.yaml",
			kptfile: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: nginx
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configPath: setters.yaml
`,
			config: `
data:
  namespace: my-space
  image: ubuntu
  list: |
    - dev
    - prod
  tag: 1.14.2
`,
			expectedKptfile: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: nginx
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configPath: setters.yaml
`,
			errMsg: `file "setters.yaml" doesn't exist, please ensure the file specified in "configPath" exists and retry`,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			baseDir, err := ioutil.TempDir("", "")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			defer os.RemoveAll(baseDir)

			err = ioutil.WriteFile(filepath.Join(baseDir, "Kptfile"), []byte(test.kptfile), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			err = ioutil.WriteFile(filepath.Join(baseDir, "setters.yaml"), []byte(test.settersConfig), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			s := &PutSetterValues{}
			node, err := kyaml.Parse(test.config)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			Decode(node, s)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			inout := &kio.LocalPackageReadWriter{
				PackagePath:     baseDir,
				NoDeleteFiles:   true,
				PackageFileName: "Kptfile",
				MatchFilesGlob:  append(kio.DefaultMatch, "Kptfile"),
			}
			err = kio.Pipeline{
				Inputs:  []kio.Reader{inout},
				Filters: []kio.Filter{s},
				Outputs: []kio.Writer{inout},
			}.Execute()
			if test.errMsg != "" {
				if !assert.NotNil(t, err) {
					t.FailNow()
				}
				if !assert.Contains(t, err.Error(), test.errMsg) {
					t.FailNow()
				}
			}

			if test.errMsg == "" && !assert.NoError(t, err) {
				t.FailNow()
			}

			actualKptfile, err := ioutil.ReadFile(filepath.Join(baseDir, "Kptfile"))
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t,
				test.expectedKptfile,
				string(actualKptfile)) {
				t.FailNow()
			}

			actualSettersConfig, err := ioutil.ReadFile(filepath.Join(baseDir, "setters.yaml"))
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t,
				test.expectedSettersConfig,
				string(actualSettersConfig)) {
				t.FailNow()
			}
		})
	}
}

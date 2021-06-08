package upsertresource

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestUpsertResourceFilter(t *testing.T) {
	var tests = []struct {
		name              string
		fnconfig          string
		input             string
		expectedResources string
		errMsg            string
	}{
		{
			name: "replace a resource",
			input: `apiVersion: v1
kind: Service
metadata:
  name: myService
  annotations:
    foo: bar
spec:
  selector:
    app: MyApp
  ports:
    - protocol: TCP
      port: 80
      targetPort: 9376
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
`,
			fnconfig: `
apiVersion: v1
kind: Service
metadata:
  name: myService
  annotations:
    abc: def
spec:
  selector:
    app: MyApp
  ports:
    - protocol: UDP
      port: 8080
      targetPort: 1234
`,
			expectedResources: `apiVersion: v1
kind: Service
metadata:
  name: myService
  annotations:
    abc: def
    config.kubernetes.io/path: f1.yaml
spec:
  selector:
    app: MyApp
  ports:
    - protocol: UDP
      port: 8080
      targetPort: 1234
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
  annotations:
    config.kubernetes.io/path: 'f1.yaml'
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: nginx
          image: nginx:1.14.2
          ports:
            - containerPort: 80
`,
		},
		{
			name: "Add a resource",
			input: `apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  selector:
    app: MyApp
  ports:
    - protocol: TCP
      port: 80
      targetPort: 9376
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
`,
			fnconfig: `
apiVersion: v1
kind: Service
metadata:
  name: myService2
spec:
  selector:
    app: MyApp
  ports:
    - protocol: UDP
      port: 8080
      targetPort: 1234
`,
			expectedResources: `apiVersion: v1
kind: Service
metadata:
  name: myService
  annotations:
    config.kubernetes.io/path: 'f1.yaml'
spec:
  selector:
    app: MyApp
  ports:
    - protocol: TCP
      port: 80
      targetPort: 9376
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
  annotations:
    config.kubernetes.io/path: 'f1.yaml'
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: nginx
          image: nginx:1.14.2
          ports:
            - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: myService2
spec:
  selector:
    app: MyApp
  ports:
    - protocol: UDP
      port: 8080
      targetPort: 1234
`,
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

			err = ioutil.WriteFile(filepath.Join(baseDir, "f1.yaml"), []byte(test.input), 0700)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			node, err := kyaml.Parse(test.fnconfig)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			s := &UpsertResource{Resource: node}
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			in := &kio.LocalPackageReader{
				PackagePath: baseDir,
			}
			out := &bytes.Buffer{}
			err = kio.Pipeline{
				Inputs:  []kio.Reader{in},
				Filters: []kio.Filter{s},
				Outputs: []kio.Writer{kio.ByteWriter{Writer: out}},
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

			if !assert.Equal(t,
				test.expectedResources,
				out.String()) {
				t.FailNow()
			}
		})
	}
}

func TestIsSameResource(t *testing.T) {
	var tests = []struct {
		name      string
		resource1 string
		resource2 string
		expected  bool
	}{
		{
			name: "same resource 1",
			resource1: `apiVersion: v1
kind: Service
metadata:
  name: myService
`,
			resource2: `apiVersion: v1
kind: Service
metadata:
  name: myService
`,
			expected: true,
		},
		{
			name: "same resource 2",
			resource1: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
`,
			resource2: `apiVersion: apps/v1alpha1
kind: Deployment
metadata:
  name: myDeployment
`,
			expected: true,
		},
		{
			name: "not same resource: default and empty namespace",
			resource1: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
`,
			resource2: `apiVersion: apps/v1alpha1
kind: Deployment
metadata:
  name: myDeployment
  namespace: default
`,
			expected: false,
		},
		{
			name: "not same resource: different kind",
			resource1: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
`,
			resource2: `apiVersion: apps/v1
kind: Service
metadata:
  name: myDeployment
`,
			expected: false,
		},
		{
			name: "not same resource: different namespace",
			resource1: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
  namespace: foo
`,
			resource2: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
  namespace: bar
`,
			expected: false,
		},
		{
			name: "not same resource: different names",
			resource1: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment1
  namespace: foo
`,
			resource2: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment2
  namespace: foo
`,
			expected: false,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {

			node1, err := kyaml.Parse(test.resource1)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			node2, err := kyaml.Parse(test.resource2)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			meta1, err := node1.GetMeta()
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			meta2, err := node2.GetMeta()
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			assert.Equal(t, test.expected, IsSameResource(meta1, meta2))
		})
	}
}

func TestCombineInputAndMatchedAnnotations(t *testing.T) {
	var tests = []struct {
		name                string
		inputResourceAnno   map[string]string
		matchedResourceAnno map[string]string
		expected            map[string]string
	}{
		{
			name: "combine annotations 1",
			inputResourceAnno: map[string]string{
				"inputFoo":                          "inputBar",
				"config.kubernetes.io/local-config": "true",
			},
			matchedResourceAnno: map[string]string{
				"existingFoo":                       "existingBar",
				"config.kubernetes.io/index":        "0",
				"config.kubernetes.io/path":         "foo.yaml",
				"config.kubernetes.io/local-config": "true",
			},
			expected: map[string]string{
				"inputFoo":                          "inputBar",
				"config.kubernetes.io/index":        "0",
				"config.kubernetes.io/path":         "foo.yaml",
				"config.kubernetes.io/local-config": "true",
			},
		},

		{
			name: "combine annotations 2",
			inputResourceAnno: map[string]string{
				"inputFoo":                          "inputBar",
				"config.kubernetes.io/local-config": "true",
			},
			matchedResourceAnno: map[string]string{
				"existingFoo":                "existingBar",
				"config.kubernetes.io/index": "0",
				"config.kubernetes.io/path":  "foo.yaml",
			},
			expected: map[string]string{
				"inputFoo":                          "inputBar",
				"config.kubernetes.io/index":        "0",
				"config.kubernetes.io/path":         "foo.yaml",
				"config.kubernetes.io/local-config": "true",
			},
		},

		{
			name: "combine annotations 3",
			inputResourceAnno: map[string]string{
				"inputFoo": "inputBar",
			},
			matchedResourceAnno: map[string]string{
				"existingFoo":                       "existingBar",
				"config.kubernetes.io/index":        "0",
				"config.kubernetes.io/path":         "foo.yaml",
				"config.kubernetes.io/local-config": "true",
			},
			expected: map[string]string{
				"inputFoo":                   "inputBar",
				"config.kubernetes.io/index": "0",
				"config.kubernetes.io/path":  "foo.yaml",
			},
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, combineInputAndMatchedAnnotations(test.inputResourceAnno, test.matchedResourceAnno))
		})
	}
}

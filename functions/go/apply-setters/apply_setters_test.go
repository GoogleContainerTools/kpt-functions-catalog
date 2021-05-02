package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestApplySettersFilter(t *testing.T) {
	var tests = []struct {
		name              string
		config            string
		input             string
		expectedResources string
		errMsg            string
	}{
		{
			name: "set name and label",
			input: `apiVersion: v1
kind: Service
metadata:
  name: myService # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: mungebot # kpt-set: ${app}
  name: mungebot
`,
			config: `
data:
  app: my-app
`,
			expectedResources: `apiVersion: v1
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
		},
		{
			name: "set name and label",
			input: `apiVersion: v1
kind: Service
metadata:
  name: myService # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: mungebot # kpt-set: ${app}
  name: mungebot
`,
			config: `
data:
  app: my-app
`,
			expectedResources: `apiVersion: v1
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
		},
		{
			name: "set setter pattern",
			config: `
data:
  image: ubuntu
  tag: 1.8.0
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: nginx
          image: nginx:1.7.9 # kpt-set: ${image}:${tag}
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: nginx
          image: ubuntu:1.8.0 # kpt-set: ${image}:${tag}
`,
		},
		{
			name: "derive missing values from pattern",
			config: `
data:
  image: ubuntu
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: nginx
          image: nginx:1.7.9 # kpt-set: ${image}:${tag}`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: nginx
          image: ubuntu:1.7.9 # kpt-set: ${image}:${tag}
`,
		},
		{
			name: "don't set if no relevant setter values are provided",
			config: `
data:
  project: my-project-foo
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: nginx
          image: nginx:1.7.9 # kpt-set: ${image}:${tag}
 `,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: nginx
          image: nginx:1.7.9 # kpt-set: ${image}:${tag}
`,
		},
		{
			name: "error if values not provided and can't be derived",
			config: `
data:
  image: ubuntu
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
        - image: irrelevant_value # kpt-set: ${image}:${tag}
          name: nginx
 `,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
        - image: irrelevant_value # kpt-set: ${image}:${tag}
          name: nginx
 `,
			errMsg: `values for setters [${tag}] must be provided`,
		},
		{
			name: "apply array setter",
			config: `
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
    - nginx
    - ubuntu
 `,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: # kpt-set: ${images}
    - ubuntu
    - hbase
`,
		},
		{
			name: "apply array setter with scalar error",
			config: `
data:
  images: ubuntu
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: # kpt-set: ${images}
    - nginx
    - ubuntu
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: # kpt-set: ${images}
    - nginx
    - ubuntu
`,
			errMsg: `input to array setter must be an array of values`,
		},
		{
			name: "apply array setter interpolation error",
			config: `
data:
  images: |
    [ubuntu, hbase]
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: # kpt-set: ${images}:${tag}
    - nginx
    - ubuntu
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: # kpt-set: ${images}:${tag}
    - nginx
    - ubuntu
`,
			errMsg: `invalid setter pattern for array node: "${images}:${tag}"`,
		},
		{
			name: "setter type handling",
			config: `
data:
  foo: false
  bar: 20
  baz: 21.22
`,
			input: `apiVersion: apps/v1
kind: MyKind
metadata:
  annotations:
    foo1: "true" # kpt-set: ${foo}
    foo2: hello # kpt-set: ${foo}
    bar: "10" # kpt-set: ${bar}
    baz: "11.12" # kpt-set: ${baz}
    bor: true-10-11.12 # kpt-set: ${foo}-${bar}-${baz}
spec:
  replicas: 10 # kpt-set: ${bar}
`,
			expectedResources: `apiVersion: apps/v1
kind: MyKind
metadata:
  annotations:
    foo1: "false" # kpt-set: ${foo}
    foo2: "false" # kpt-set: ${foo}
    bar: "20" # kpt-set: ${bar}
    baz: "21.22" # kpt-set: ${baz}
    bor: false-20-21.22 # kpt-set: ${foo}-${bar}-${baz}
spec:
  replicas: 20 # kpt-set: ${bar}
`,
		},
		{
			name: "error for incorrect input type",
			config: `
data:
  replicas: two
`,
			input: `apiVersion: apps/v1
kind: MyKind
metadata:
spec:
  replicas: 10 # kpt-set: ${replicas}
`,
			expectedResources: `apiVersion: apps/v1
kind: MyKind
metadata:
spec:
  replicas: 10 # kpt-set: ${replicas}
`,
			errMsg: `input value "two" doesn't conform to node type "int"`,
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

			r, err := ioutil.TempFile(baseDir, "k8s-cli-*.yaml")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			defer os.Remove(r.Name())
			err = ioutil.WriteFile(r.Name(), []byte(test.input), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			s := &ApplySetters{}
			node, err := kyaml.Parse(test.config)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			err = decode(node, s)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			inout := &kio.LocalPackageReadWriter{
				PackagePath:     baseDir,
				NoDeleteFiles:   true,
				PackageFileName: "Kptfile",
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

			actualResources, err := ioutil.ReadFile(r.Name())
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t,
				test.expectedResources,
				string(actualResources)) {
				t.FailNow()
			}
		})
	}
}

type patternTest struct {
	name     string
	value    string
	pattern  string
	expected map[string]string
}

var resolvePatternCases = []patternTest{
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
		name:    "setter values from pattern 1",
		value:   "nginx:1.7.1",
		pattern: `${image}:${tag}`,
		expected: map[string]string{
			"image": "nginx",
			"tag":   "1.7.1",
		},
	},
	{
		name:     "setter values from pattern unresolved",
		value:    "foo-dev-bar-us-east-1-baz",
		pattern:  `${image}:${tag}`,
		expected: map[string]string{},
	},
}

func TestCurrentSetterValues(t *testing.T) {
	for _, tests := range [][]patternTest{resolvePatternCases} {
		for i := range tests {
			test := tests[i]
			t.Run(test.name, func(t *testing.T) {
				res := currentSetterValues(test.pattern, test.value)
				if !assert.Equal(t, test.expected, res) {
					t.FailNow()
				}
			})
		}
	}
}

type typeTest struct {
	name    string
	yamlTag string
	value   string
	isError bool
}

var validateTypeCases = []typeTest{
	{
		name:    "string can't be bool",
		value:   "foo",
		yamlTag: kyaml.NodeTagBool,
		isError: true,
	},
	{
		name:    "string can't be float",
		value:   "foo",
		yamlTag: kyaml.NodeTagFloat,
		isError: true,
	},
	{
		name:    "string can't be int",
		value:   "foo",
		yamlTag: kyaml.NodeTagInt,
		isError: true,
	},
	{
		name:    "string can't be sequence",
		value:   "foo",
		yamlTag: kyaml.NodeTagSeq,
		isError: true,
	},
	{
		name:    "string can't be map",
		value:   "foo",
		yamlTag: kyaml.NodeTagMap,
		isError: true,
	},
	{
		name:    "bool can't be int",
		value:   "true",
		yamlTag: kyaml.NodeTagInt,
		isError: true,
	},
	{
		name:    "int can't be float",
		value:   "10",
		yamlTag: kyaml.NodeTagFloat,
		isError: true,
	},
	{
		name:    "float can't be int",
		value:   "10.1",
		yamlTag: kyaml.NodeTagInt,
		isError: true,
	},
	{
		name:    "bool can be string",
		value:   "true",
		yamlTag: kyaml.NodeTagString,
		isError: false,
	},
	{
		name:    "float can be string",
		value:   "1.22",
		yamlTag: kyaml.NodeTagString,
		isError: false,
	},
	{
		name:    "int can be string",
		value:   "1",
		yamlTag: kyaml.NodeTagString,
		isError: false,
	},
	{
		name:    "sequence can be string",
		value:   "[foo, bar]",
		yamlTag: kyaml.NodeTagString,
		isError: false,
	},
	{
		name:    "map can be string",
		value:   "a: b",
		yamlTag: kyaml.NodeTagString,
		isError: false,
	},
}

func TestValidateType(t *testing.T) {
	for _, tests := range [][]typeTest{validateTypeCases} {
		for i := range tests {
			test := tests[i]
			t.Run(test.name, func(t *testing.T) {
				err := validateType(test.yamlTag, test.value)
				if test.isError {
					if !assert.Error(t, err) {
						t.FailNow()
					}
				} else {
					if !assert.NoError(t, err) {
						t.FailNow()
					}
				}
			})
		}
	}
}

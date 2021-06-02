package createsetters

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestCreateSettersFilter(t *testing.T) {
	var tests = []struct {
		name              string
		config            string
		input             string
		expectedResources string
		errMsg            string
	}{
		{
			name: "set comment for array setter of flow style",
			config: `
data:
  env: |
    [foo, bar]
  name: nginx
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  env: [foo, bar]
 `,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment # kpt-set: ${name}-deployment
  env: # kpt-set: ${env}
    - foo
    - bar
`,
		},
		{
			name: "set comment for scalar nodes",
			config: `
data:
  name: nginx
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  env: [foo, bar]
 `,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment # kpt-set: ${name}-deployment
  env: [foo, bar]
`,
		},
		{
			name: "set comment for scalar and sequence nodes",
			input: `apiVersion: v1
kind: Service
metadata:
  name: myService
  namespace: foo 
image: nginx:1.7.1 
env: [foo, bar] 
`,
			config: `
data:
  app: myService
  ns: foo
  image: nginx
  tag: 1.7.1
  env: "[foo, bar]"
`,
			expectedResources: `apiVersion: v1
kind: Service
metadata:
  name: myService # kpt-set: ${app}
  namespace: foo # kpt-set: ${ns}
image: nginx:1.7.1 # kpt-set: ${image}:${tag}
env: # kpt-set: ${env}
  - foo # kpt-set: ${ns}
  - bar
`,
		},

		{
			name: "scalar setter donot match",
			config: `
data:
  name: ubuntu
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  env: [foo, bar]
 `,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  env: [foo, bar]
`,
		},
		{
			name: "array setter with flow style donot match",
			config: `
data:
  env: |
    [foo, bar, pro]
  name: nginx
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  env: [foo, bar]
 `,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment # kpt-set: ${name}-deployment
  env: [foo, bar]
`,
		},
		{
			name: "setter with no matching values",
			config: `
data:
  app: myService
  ns: foo
  images: |
    - ubuntu
    - linux
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images:
    - nginx
    - ubuntu
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images:
    - nginx
    - ubuntu
`,
		},
		{
			name: "containing overlap values",
			config: `
data:
  image: nginx
  name: image
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment # kpt-set: ${image}-deployment
`,
		},
		{
			name: "FlowStyle to FoldedStyle",
			config: `
data:
  image: "[nginx, ubuntu]"
  os: ubuntu
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: [nginx, ubuntu]
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: # kpt-set: ${image}
    - nginx
    - ubuntu # kpt-set: ${os}
`,
		},
		{
			name: "Empty array values",
			config: `
data:
  image: "[]"
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: []
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: [] # kpt-set: ${image}
`,
		},
		{
			name: "Empty data map",
			config: `
data: {}
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: []
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: []
`,
			errMsg: "config map cannot be empty",
		},

		{
			name: "some values in sequence node match",
			config: `
data:
  app: myService
  image: nginx
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: [nginx, ubuntu]
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment # kpt-set: ${image}-deployment
spec:
  images:
    - nginx # kpt-set: ${image}
    - ubuntu
`,
		},

		{
			name: "scalar setter donot match",
			config: `
data:
  name: ubuntu
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  env: [foo, bar]
 `,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  env: [foo, bar]
`,
		},
		{
			name: "array setter with flow style donot match",
			config: `
data:
  env: |
    [foo, bar, pro]
  name: nginx
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  env: [foo, bar]
 `,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment # kpt-set: ${name}-deployment
  env: [foo, bar]
`,
		},
		{
			name: "create array setter with scalar error",
			config: `
data:
  app: myService
  ns: foo
  images: |
    - ubuntu
    - linux
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images:
    - nginx
    - ubuntu
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images:
    - nginx
    - ubuntu
`,
		},
		{
			name: "containing overlap values",
			config: `
data:
  image: nginx
  name: image
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment # kpt-set: ${image}-deployment
`,
		},
		{
			name: "FlowStyle to FoldedStyle",
			config: `
data:
  image: "[nginx, ubuntu]"
  os: ubuntu
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: [nginx, ubuntu]
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: # kpt-set: ${image}
    - nginx
    - ubuntu # kpt-set: ${os}
`,
		},
		{
			name: "Empty array values",
			config: `
data:
  image: []
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: [nginx, ubuntu]
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: [nginx, ubuntu]
`,
			errMsg: "input setters list cannot be empty",
		},

		{
			name: "create array setter with scalar error",
			config: `
data:
  app: myService
  image: nginx
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: [nginx, ubuntu]
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment # kpt-set: ${image}-deployment
spec:
  images:
    - nginx # kpt-set: ${image}
    - ubuntu
`,
			errMsg: `parsing error in Config Map`,
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

			s := &CreateSetters{}
			node, err := kyaml.Parse(test.config)
			if err != nil {
				err = fmt.Errorf("parsing error in Config Map")
				if test.errMsg != "" && !assert.Equal(t, err.Error(), test.errMsg) {
					t.FailNow()
				} else if test.errMsg == "" {
					t.FailNow()
				}
				return
			}
			err = Decode(node, s)
			if err != nil {
				if test.errMsg != "" && !assert.Equal(t, err.Error(), test.errMsg) {
					t.FailNow()
				} else if test.errMsg == "" {
					t.FailNow()
				}
				return
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

type lineCommentTest struct {
	name    string
	value   string
	comment string
}

var resolveLineCommentCases = []lineCommentTest{
	{
		name:    "comment for pattern 1",
		value:   "foo-dev-bar-us-east-baz",
		comment: `foo-${role}-bar-${region}-baz`,
	},
	{
		name:    "comment for pattern 2",
		value:   "nginx:1.2.1",
		comment: `${image}:1.2.1`,
	},
	{
		name:    "comment for pattern 3",
		value:   "nginx:1.1.2",
		comment: `${image}:${tag}`,
	},
	{
		name:    "comment for pattern 4",
		value:   "ubuntu",
		comment: `${env}`,
	},
	{
		name:    "comment for pattern 5",
		value:   "linux",
		comment: ``,
	},
}

var inputSetters = []ScalarSetter{
	{
		Name:  "env",
		Value: "ubuntu",
	},
	{
		Name:  "image",
		Value: "nginx",
	},
	{
		Name:  "role",
		Value: "dev",
	},
	{
		Name:  "tag",
		Value: "1.1.2",
	},
	{
		Name:  "region",
		Value: "us-east",
	},
}

func TestCurrentSetterValues(t *testing.T) {
	for _, tests := range [][]lineCommentTest{resolveLineCommentCases} {
		for i := range tests {
			test := tests[i]
			t.Run(test.name, func(t *testing.T) {
				res, match := getLineComment(test.value, inputSetters)
				if match {
					if !assert.Equal(t, test.comment, res) {
						t.FailNow()
					}
				} else {
					if !assert.Equal(t, test.comment, "") {
						t.FailNow()
					}
				}

			})
		}
	}
}

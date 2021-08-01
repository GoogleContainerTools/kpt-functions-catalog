package injector

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestConfigMapInjector(t *testing.T) {
	var tests = []struct {
		name              string
		config            string
		input             string
		expectedResources string
		errMsg            string
	}{
		{
			name: "inject config",
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: fn-config-my-app-values
  annotations:
    local-config: "true"
data:
  issuer-url: https://auth.dev.example.com/auth/realms/my-app # kpt-set: https://auth.${stage}.${domain}/auth/realms/${realm}
  realm: my-realm # kpt-set: ${realm}
  s3-url: https://s3-server.com/bucket # kpt-set: ${s3-base-url}/${s3-bucket}
  python-version: python2
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: fn-config-my-app-template
  annotations:
    config.kubernetes.io/local-config: "true"
data:
  oidc.config: |
    name: my-app
    clientID: $oidc.client-id
    clientSecret: $oidc.client-secret
    issuer: ${issuer-url}
    requestedScopes:
    - openid
    - profile
    - email
    - groups
  config.json: |
    {
      "deployment": {
        "files": {
          "example-resource-file1": {
            "sourceUrl": "${s3-url}/example-application/example-resource-file1"
          },
          "images/example-resource-file2": {
            "sourceUrl": "${s3-url}/example-application/images/example-resource-file2"
          },
        }
      },
      "id": "v1",
      "runtime": "${python-version}",
      "threadsafe": true,
    }
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: my-app-cm
data:
  adminEnabled: "true"
  url: https://my-app.dev.example.com # kpt-set: https://my-app.${stage}.${domain}
`,
			config: `
apiVersion: fn.kumorilabs.io/v1alpha1
kind: ConfigMapInjector
metadata:
  name: my-config-injector
spec:
  target:
    name: my-app-cm
  values:
    name: fn-config-my-app-values
  template:
    name: fn-config-my-app-template
`,
			expectedResources: `apiVersion: v1
kind: ConfigMap
metadata:
  name: fn-config-my-app-values
  annotations:
    local-config: "true"
data:
  issuer-url: https://auth.dev.example.com/auth/realms/my-app # kpt-set: https://auth.${stage}.${domain}/auth/realms/${realm}
  realm: my-realm # kpt-set: ${realm}
  s3-url: https://s3-server.com/bucket # kpt-set: ${s3-base-url}/${s3-bucket}
  python-version: python2
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: fn-config-my-app-template
  annotations:
    config.kubernetes.io/local-config: "true"
data:
  oidc.config: |
    name: my-app
    clientID: $oidc.client-id
    clientSecret: $oidc.client-secret
    issuer: ${issuer-url}
    requestedScopes:
    - openid
    - profile
    - email
    - groups
  config.json: |
    {
      "deployment": {
        "files": {
          "example-resource-file1": {
            "sourceUrl": "${s3-url}/example-application/example-resource-file1"
          },
          "images/example-resource-file2": {
            "sourceUrl": "${s3-url}/example-application/images/example-resource-file2"
          },
        }
      },
      "id": "v1",
      "runtime": "${python-version}",
      "threadsafe": true,
    }
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: my-app-cm
data:
  adminEnabled: "true"
  config.json: |
    {
      "deployment": {
        "files": {
          "example-resource-file1": {
            "sourceUrl": "https://s3-server.com/bucket/example-application/example-resource-file1"
          },
          "images/example-resource-file2": {
            "sourceUrl": "https://s3-server.com/bucket/example-application/images/example-resource-file2"
          },
        }
      },
      "id": "v1",
      "runtime": "python2",
      "threadsafe": true,
    }
  oidc.config: |
    name: my-app
    clientID: $oidc.client-id
    clientSecret: $oidc.client-secret
    issuer: https://auth.dev.example.com/auth/realms/my-app
    requestedScopes:
    - openid
    - profile
    - email
    - groups
  url: https://my-app.dev.example.com
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

			r, err := ioutil.TempFile(baseDir, "k8s-cli-*.yaml")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			defer os.Remove(r.Name())
			err = ioutil.WriteFile(r.Name(), []byte(test.input), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			s := &ConfigMapInjector{}
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
			}

			pipe := &kio.Pipeline{
				Inputs:  []kio.Reader{inout},
				Filters: []kio.Filter{kio.FilterFunc(s.Filter)},
				Outputs: []kio.Writer{inout},
			}

			err = pipe.Execute()
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

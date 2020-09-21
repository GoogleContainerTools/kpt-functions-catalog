package function

import (
	"testing"

	"bytes"
	"os"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestTemplates(t *testing.T) {

	os.Setenv("TESTTEMPLATE", "testtemplatevalue")

	tc := []struct {
		cfg         string
		in          string
		expectedOut string
		expectedErr bool
	}{
		{
			cfg: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: notImportantHere
data:
  template: |
    value1: {{ .literal1 }}
  literal1: value1
  literal2: value2
`,
			expectedOut: `value1: value1
`,
		},
		{
			cfg: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: notImportantHere
data:
  template: 'value: {{ env "TESTTEMPLATE" }}'
`,
			expectedOut: `value: testtemplatevalue
`,
		},
		{
			cfg: `
apiVersion: v1alpha1
kind: Templater
metadata:
  name: notImportantHere
data:
  template: |
    {{ range .hosts -}}
    ---
    apiVersion: metal3.io/v1alpha1
    kind: BareMetalHost
    metadata:
      name: {{ .name }}
    spec:
      bootMACAddress: {{ .macAddress }}
    {{ end -}}
  hosts:
    - macAddress: 00:aa:bb:cc:dd
      name: node-1
    - macAddress: 00:aa:bb:cc:ee
      name: node-2
`,
			expectedOut: `apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: node-1
spec:
  bootMACAddress: 00:aa:bb:cc:dd
---
apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: node-2
spec:
  bootMACAddress: 00:aa:bb:cc:ee
`,
		},
		{
			cfg: `
apiVersion: v1alpha1
kind: Templater
metadata:
  name: notImportantHere
data:
  template: '{{ toYaml . -}}'

  test:
    of:
      - toYaml
`,
			expectedOut: `test:
  of:
  - toYaml
`,
		},
		{
			cfg: `
apiVersion: v1alpha1
kind: Templater
metadata:
  name: notImportantHere
data:
  template: |
    {{ toYaml ignorethisbadinput -}}
  test:
    of:
      - badToYamlInput
`,
			expectedErr: true,
		},
		{
			cfg: `
apiVersion: v1alpha1
kind: Templater
metadata:
  name: notImportantHere
data:
  template: |
    {{ end }
`,
			expectedErr: true,
		},
		{
			cfg: `
apiVersion: v1alpha1
kind: Templater
metadata:
  name: notImportantHere
data:
`,
			expectedErr: true,
		},
		{
			cfg: `
apiVersion: v1alpha1
kind: Templater
metadata:
  name: notImportantHere
data:
  template: 234
`,
			expectedErr: true,
		},
		{
			cfg: `
apiVersion: v1alpha1
kind: Templater
metadata:
  name: notImportantHere
data:
  template:
    x:
      y: z
`,
			expectedErr: true,
		},
	}

	for i, ti := range tc {
		fcfg := Config{}
		err := yaml.Unmarshal([]byte(ti.cfg), &fcfg)
		if err != nil {
			t.Errorf("can't unmarshal config %s: %v. continue", ti.cfg, err)
			continue
		}

		nodes, err := (&kio.ByteReader{Reader: bytes.NewBufferString(ti.in)}).Read()
		if err != nil {
			t.Errorf("can't unmarshal in yamls %s: %v. continue", ti.in, err)
			continue
		}

		f, err := NewFilter(&fcfg)
		if err != nil {
			if !ti.expectedErr {
				t.Errorf("can't create filter for config %s: %v. continue", ti.in, err)
			}
			continue
		}
		nodes, err = f.Filter(nodes)
		if err != nil && !ti.expectedErr {
			t.Errorf("exec %d returned unexpected error %v for %s", i, err, ti.cfg)
			continue
		}
		out := &bytes.Buffer{}
		err = kio.ByteWriter{Writer: out}.Write(nodes)
		if err != nil {
			t.Errorf("write returned unexpected error %v for %s", err, ti.cfg)
			continue
		}
		if out.String() != ti.expectedOut {
			t.Errorf("expected %s, got %s", ti.expectedOut, out.String())
		}
	}

}

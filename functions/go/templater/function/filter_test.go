package function

import (
	"testing"

	"bytes"
	"os"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestTemplates(t *testing.T) {

	os.Setenv("TESTENV", "testenvvalue")

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
  entrypoint: |
    value1: {{ .Data.literal1 }}
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
  entrypoint: 'value: {{ env "TESTENV" }}'
`,
			expectedOut: `value: testenvvalue
`,
		},
		{
			cfg: `
apiVersion: v1alpha1
kind: Templater
metadata:
  name: notImportantHere
data:
  entrypoint: |
    {{ range .Data.hosts -}}
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
  entrypoint: '{{ toYaml .Data -}}'

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
  entrypoint: |
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
  entrypoint: |
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
  entrypoint: 234
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
  entrypoint:
    x:
      y: z
`,
			expectedErr: true,
		},
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
`,
			cfg: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: notImportantHere
data:
  entrypoint: |
    {{- $_ := (set . "Items" (list)) -}}
    value: {{ env "TESTENV" }}
`,
			expectedOut: `value: testenvvalue
`,
		},
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
`,
			cfg: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: notImportantHere
data:
  entrypoint: |
    value: {{ env "TESTENV" }}
`,
			expectedOut: `apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
---
value: testenvvalue
`,
		},
		// transformer tests
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
`,
			cfg: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: notImportantHere
data:
  annotationTransf: |
    kind: AnnotationSetter
    key: test-annotation
    value: %s
  entrypoint: |
    {{- $_ := KPipe .Items (list (KYFilter (list (YFilter (printf .Data.annotationTransf (env "TESTENV")))))) -}}
`,
			expectedOut: `apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
  annotations:
    test-annotation: 'testenvvalue'
`,
		},
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
data:
  value: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: map2
data:
  value: value2
`,
			cfg: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: notImportantHere
data:
  map1grep: |
    kind: GrepFilter
    path: 
    - metadata
    - name
    value: ^map1$
  pathGet1: |
    kind: PathGetter
    path:
    - data
    - value
  map2grep: |
    kind: GrepFilter
    path:
    - metadata
    - name
    value: ^map2$
  map2PathGet: |
    kind: PathGetter
    path:
    - data
  fieldSet: |
    kind: FieldSetter
    name: value
    stringValue: %s
  entrypoint: |
    {{- $map1 := KPipe .Items (list (KFilter .Data.map1grep)) -}}
    {{- $map1value := YValue (YPipe (index $map1 0) (list (YFilter .Data.pathGet1))) -}}
    {{- $_ := KPipe .Items (list (KFilter .Data.map2grep) (KYFilter (list (YFilter .Data.map2PathGet) (YFilter (printf .Data.fieldSet $map1value))))) -}}
`,
			expectedOut: `apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
data:
  value: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: map2
data:
  value: value1
`,
		},
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
  annotations:
    test-annotation: x
data:
  value: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: map2
data:
  value: value2
`,
			cfg: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: notImportantHere
data:
  grep: |
    kind: GrepFilter
    path:
    - metadata
    - annotations
    - test-annotation
    value: ^x$
    invertMatch: true
  entrypoint: |
    {{- $_ := (set . "Items" (KPipe .Items (list (KFilter .Data.grep)))) -}}
`,
			expectedOut: `apiVersion: v1
kind: ConfigMap
metadata:
  name: map2
data:
  value: value2
`,
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

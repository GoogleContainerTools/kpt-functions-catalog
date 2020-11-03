package function

import (
	"testing"

	"bytes"
	"os"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestTemplates(t *testing.T) {

	os.Setenv("TESTENV1", "testenvvalue1")
	os.Setenv("TESTENV2", "testenvvalue2")

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
  literal1: value1
  literal2: value2
`,
			in: `apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
`,
			expectedOut: `apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
  annotations:
    literal1: 'value1'
    literal2: 'value2'
`,
		},
		{
			cfg: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: notImportantHere
data:
  TESTENV1: ''
  TESTENV2: ''
`,
			in: `apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
`,
			expectedOut: `apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
  annotations:
    TESTENV1: 'testenvvalue1'
    TESTENV2: 'testenvvalue2'
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

package searchreplace

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

type test struct {
	name              string
	config            string
	input             string
	expectedResources string
	out               string
	errMsg            string
}

func TestSearchCommand(t *testing.T) {
	for _, tests := range [][]test{searchReplaceCases, putPatternCases} {
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

				s := &SearchReplace{}
				node, err := kyaml.Parse(test.config)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				inout := &kio.LocalPackageReadWriter{
					PackagePath:     baseDir,
					NoDeleteFiles:   true,
					PackageFileName: "Kptfile",
				}
				err = Decode(node, s)
				if !assert.NoError(t, err) {
					t.FailNow()
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
					return
				}

				if test.errMsg == "" && !assert.NoError(t, err) {
					t.FailNow()
				}

				actualResources, err := ioutil.ReadFile(r.Name())
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t,
					strings.TrimSpace(test.expectedResources),
					strings.TrimSpace(string(actualResources))) {
					t.FailNow()
				}

				expectedOut := strings.ReplaceAll(test.out, "${filePath}", filepath.Base(r.Name()))
				if !assert.Equal(t, expectedOut, s.resultsString()) {
					t.FailNow()
				}
			})
		}
	}
}

func TestDecode(t *testing.T) {
	rn, err := kyaml.Parse(`data:
  by-value: foo
  put-values: bar`)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = Decode(rn, &SearchReplace{})
	if !assert.Error(t, err) {
		t.FailNow()
	}
	expected := `invalid matcher "put-values", must be one of ["by-value" "by-value-regex" "by-path" "put-value" "put-comment"]`
	if !assert.Equal(t, expected, err.Error()) {
		t.FailNow()
	}

}

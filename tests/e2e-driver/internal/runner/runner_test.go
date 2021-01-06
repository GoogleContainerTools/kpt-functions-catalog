package runner_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/GoogleContainerTools/kpt-functions-catalog/tests/e2e-driver/internal/runner"
	"github.com/stretchr/testify/assert"
)

func TestNewRunner(t *testing.T) {
	_, err := NewRunner("non-exist", false)
	assert.Error(t, err, "expect an error")

	_, err = NewRunner("/", false)
	assert.NoError(t, err, "unexpect error")
}

func TestRunnerRun(t *testing.T) {
	configYaml := `configs:
- pkgPath: ../../../../examples/set-namespace
- pkgPath: ../../../../examples/kubeval
  network: true`

	d, err := ioutil.TempDir("", "kpt-fn-catalog-e2e-driver-unit-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(d)
	configPath := filepath.Join(d, "config.yaml")
	err = ioutil.WriteFile(configPath, []byte(configYaml), 0644)
	assert.NoError(t, err)
	config, err := ConfigFromFile(configPath)
	assert.NoError(t, err)
	for _, c := range config.Configs {
		r, err := NewRunner(c.PkgPath, c.Network)
		assert.NoError(t, err)
		err = r.Run()
		assert.NoError(t, err)
	}
}

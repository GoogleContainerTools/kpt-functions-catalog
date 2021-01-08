package runner_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/GoogleContainerTools/kpt-functions-catalog/tests/e2etest/internal/runner"
	"github.com/stretchr/testify/assert"
)

func TestConfigFromFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "kpt-fn-catalog-e2e-driver-unit-test-*")
	assert.NoError(t, err, "failed to create temporary directory")
	defer os.RemoveAll(tmpDir)
	config := `
configs:
- pkgPath: a/b
  network: true
- pkgPath: foo
- pkgPath: bar
  network: false
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	err = ioutil.WriteFile(configPath, []byte(config), 0644)
	assert.NoError(t, err, "failed to write to config file")
	expected := TestConfigs{
		Configs: []TestConfig{
			TestConfig{
				PkgPath: "a/b",
				Network: true,
			},
			TestConfig{
				PkgPath: "foo",
				Network: false,
			},
			TestConfig{
				PkgPath: "bar",
				Network: false,
			},
		},
	}
	actual, err := ConfigFromFile(configPath)
	assert.NoError(t, err, "failed to read from config file")
	if !assert.EqualValues(t, expected, *actual, "actual doesn't match expected") {
		t.FailNow()
	}
}

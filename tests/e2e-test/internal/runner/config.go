package runner

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// TestConfig contains the information needed to run a test. Each test case
// run by this driver is described by a `TestConfig`.
//
// Example of a test config:
//
// 	pkgPath: foo/bar
//	network: true
type TestConfig struct {
	// PkgPath is the path to the package which will be tested.
	PkgPath string `json:"pkgPath,omitempty" yaml:"pkgPath,omitempty"`
	// Network indicates whether enable network access for function or not
	Network bool `json:"network,omitempty" yaml:"network,omitempty"`
}

// TestConfigs contains a list of TestConfig. These configs should be read
// from a config YAML file.
//
// Example of a config file
//
// 	configs:
// 	- pkgPath: ../examples/set-namespace
// 	- pkgPath: ../examples/kubeval
// 	  network: true
type TestConfigs struct {
	Configs []TestConfig `json:"configs,omitempty" yaml:"configs,omitempty"`
}

// ConfigFromFile returns a list of TestConfig read from path
func ConfigFromFile(path string) (*TestConfigs, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open test config: %w", err)
	}
	var configs TestConfigs
	err = yaml.Unmarshal(b, &configs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return &configs, nil
}

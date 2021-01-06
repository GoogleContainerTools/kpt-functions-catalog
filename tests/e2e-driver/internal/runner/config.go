package runner

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// TestConfig is the config for a test
type TestConfig struct {
	PkgPath string `json:"pkgPath,omitempty" yaml:"pkgPath,omitempty"`
	Network bool   `json:"network,omitempty" yaml:"network,omitempty"`
}

// TestConfigs contains a list of TestConfig
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

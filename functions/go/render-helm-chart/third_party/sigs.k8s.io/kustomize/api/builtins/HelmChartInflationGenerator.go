// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate HelmChartInflationGeneratorPluginator
package builtins

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/render-helm-chart/third_party/sigs.k8s.io/kustomize/api/types"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/hasher"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/yaml"
)

// Add the given labels to the given field specifications.
type HelmArgs struct {
	types.HelmGlobals `json:"helmGlobals,omitempty" yaml:"helmGlobals,omitempty"`
	HelmCharts []types.HelmChart `json:"helmCharts,omitempty" yaml:"helmCharts,omitempty"`
}

type HelmChartInflationGeneratorPlugin struct {
	types.HelmGlobals `json:",inline,omitempty" yaml:",inline,omitempty"`
	types.HelmChart `json:",inline,omitempty" yaml:",inline,omitempty"`
	tmpDir string
}

const (
	valuesMergeOptionMerge    = "merge"
	valuesMergeOptionOverride = "override"
	valuesMergeOptionReplace  = "replace"
)

var legalMergeOptions = []string{
	valuesMergeOptionMerge,
	valuesMergeOptionOverride,
	valuesMergeOptionReplace,
}

//noinspection GoUnusedGlobalVariable
var KustomizeHelmChartInflationGeneratorPlugin HelmChartInflationGeneratorPlugin

func (p *HelmChartInflationGeneratorPlugin) establishTmpDir() (err error) {
	if p.tmpDir != "" {
		// already done.
		return nil
	}
	p.tmpDir, err = ioutil.TempDir("", "kustomize-helm-")
	return err
}

func (p *HelmChartInflationGeneratorPlugin) ValidateArgs() (err error) {
	if p.Name == "" {
		return fmt.Errorf("chart name cannot be empty")
	}

	// ChartHome might be written to by the function in a container,
	// so it must be under the `/tmp` directory
	if p.ChartHome == "" {
		p.ChartHome = "tmp/charts"
	}

	if p.ValuesFile == "" {
		p.ValuesFile = filepath.Join(p.ChartHome, p.Name, "values.yaml")
	}


	if err = p.errIfIllegalValuesMerge(); err != nil {
		return err
	}

	if p.IncludeCRDs == "" {
		p.IncludeCRDs = "false"
	}

	if p.IncludeCRDs != "true" && p.IncludeCRDs != "false" {
		return fmt.Errorf("includeCRDs must be 'true' or 'false'")
	}

	// ConfigHome is not loaded by the HelmChartInflationGeneratorPlugin, and can be located anywhere.
	if p.ConfigHome == "" {
		if err = p.establishTmpDir(); err != nil {
			return errors.Wrap(
				err, "unable to create tmp dir for HELM_CONFIG_HOME")
		}
		p.ConfigHome = filepath.Join(p.tmpDir, "helm")
	}
	return nil
}

func (p *HelmChartInflationGeneratorPlugin) errIfIllegalValuesMerge() error {
	if p.ValuesMerge == "" {
		// Use the default.
		p.ValuesMerge = valuesMergeOptionOverride
		return nil
	}
	for _, opt := range legalMergeOptions {
		if p.ValuesMerge == opt {
			return nil
		}
	}
	return fmt.Errorf("valuesMerge must be one of %v", legalMergeOptions)
}

func (p *HelmChartInflationGeneratorPlugin) absChartHome() string {
	path, _ := filepath.Abs(p.ChartHome)
	return path
}

func (p *HelmChartInflationGeneratorPlugin) runHelmCommand(
	args []string) ([]byte, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := exec.Command("helm", args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	env := []string{
		fmt.Sprintf("HELM_CONFIG_HOME=%s", p.ConfigHome),
		fmt.Sprintf("HELM_CACHE_HOME=%s/.cache", p.ConfigHome),
		fmt.Sprintf("HELM_DATA_HOME=%s/.data", p.ConfigHome)}
	cmd.Env = append(os.Environ(), env...)
	err := cmd.Run()
	if err != nil {
		helm := "helm"
		err = errors.Wrap(
			fmt.Errorf(
				"unable to run: '%s %s' with env=%s (is '%s' installed?)",
				helm, strings.Join(args, " "), env, helm),
			stderr.String(),
		)
	}
	return stdout.Bytes(), err
}

// createNewMergedValuesFile replaces/merges original values file with ValuesInline.
func (p *HelmChartInflationGeneratorPlugin) createNewMergedValuesFile() (
	path string, err error) {
	if p.ValuesMerge == valuesMergeOptionMerge ||
		p.ValuesMerge == valuesMergeOptionOverride {
		if err = p.replaceValuesInline(); err != nil {
			return "", err
		}
	}
	var b []byte
	b, err = yaml.Marshal(p.ValuesInline)
	if err != nil {
		return "", err
	}
	return p.writeValuesBytes(b)
}

func (p *HelmChartInflationGeneratorPlugin) replaceValuesInline() error {
	pValues, err := ioutil.ReadFile(p.ValuesFile)
	if err != nil {
		return err
	}
	chValues := make(map[string]interface{})
	if err = yaml.Unmarshal(pValues, &chValues); err != nil {
		return err
	}
	switch p.ValuesMerge {
	case valuesMergeOptionOverride:
		err = mergo.Merge(
			&chValues, p.ValuesInline, mergo.WithOverride)
	case valuesMergeOptionMerge:
		err = mergo.Merge(&chValues, p.ValuesInline)
	}
	p.ValuesInline = chValues
	return err
}

// copyValuesFile to avoid branching.  TODO: get rid of this.
func (p *HelmChartInflationGeneratorPlugin) copyValuesFile() (string, error) {
	var b []byte
	var err error
	b, err = ioutil.ReadFile(p.ValuesFile)
	if err != nil {
		path := p.ValuesFile
		if u, err := url.Parse(path); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
			var hc *http.Client
			hc = &http.Client{}

			resp, err := hc.Get(path)
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()
			b, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}
		}
	}
	return p.writeValuesBytes(b)
}

// Write a absolute path file in the tmp file system.
func (p *HelmChartInflationGeneratorPlugin) writeValuesBytes(
	b []byte) (string, error) {
	if err := p.establishTmpDir(); err != nil {
		return "", fmt.Errorf("cannot create tmp dir to write helm values")
	}
	path := filepath.Join(p.tmpDir, p.Name+"-kustomize-values.yaml")
	return path, ioutil.WriteFile(path, b, 0644)
}

func (p *HelmChartInflationGeneratorPlugin) cleanup() {
	if p.tmpDir != "" {
		os.RemoveAll(p.tmpDir)
	}
}

func (p *HelmChartInflationGeneratorPlugin) Generate() (rm resmap.ResMap, err error) {
	defer p.cleanup()
	if err = p.checkHelmVersion(); err != nil {
		return nil, err
	}
	if path, exists := p.chartExistsLocally(); !exists {
		if p.Repo == "" {
			return nil, fmt.Errorf(
				"no repo specified for pull, no chart found at '%s'", path)
		}
		if _, err := p.runHelmCommand(p.pullCommand()); err != nil {
			return nil, err
		}
	}
	if len(p.ValuesInline) > 0 {
		p.ValuesFile, err = p.createNewMergedValuesFile()
	} else {
		p.ValuesFile, err = p.copyValuesFile()
	}
	if err != nil {
		return nil, err
	}
	var stdout []byte
	stdout, err = p.runHelmCommand(p.templateCommand())
	if err != nil {
		return nil, err
	}

	factory := NewResMapFactory()
	rm, err = factory.NewResMapFromBytes(stdout)
	if err == nil {
		return rm, nil
	}
	// try to remove the contents before first "---" because
	// helm may produce messages to stdout before it
	stdoutStr := string(stdout)
	if idx := strings.Index(stdoutStr, "---"); idx != -1 {
		return factory.NewResMapFromBytes([]byte(stdoutStr[idx:]))
	}
	return nil, err
}

func (p *HelmChartInflationGeneratorPlugin) templateCommand() []string {
	args := []string{"template"}
	if p.ReleaseName != "" {
		args = append(args, p.ReleaseName)
	}
	if p.Namespace != "" {
		args = append(args, "--namespace", p.Namespace)
	}
	args = append(args, filepath.Join(p.absChartHome(), p.Name))
	if p.ValuesFile != "" {
		args = append(args, "--values", p.ValuesFile)
	}
	if p.ReleaseName == "" {
		// AFAICT, this doesn't work as intended due to a bug in helm.
		// See https://github.com/helm/helm/issues/6019
		// I've tried placing the flag before and after the name argument.
		args = append(args, "--generate-name")
	}
	if p.IncludeCRDs == "true" {
		args = append(args, "--include-crds")
	}
	return args
}

func (p *HelmChartInflationGeneratorPlugin) pullCommand() []string {
	args := []string{
		"pull",
		"--untar",
		"--untardir", p.absChartHome(),
		"--repo", p.Repo,
		p.Name}
	if p.Version != "" {
		args = append(args, "--version", p.Version)
	}
	return args
}

// chartExistsLocally will return true if the chart does exist in
// local chart home.
func (p *HelmChartInflationGeneratorPlugin) chartExistsLocally() (string, bool) {
	path := filepath.Join(p.absChartHome(), p.Name)
	s, err := os.Stat(path)
	if err != nil {
		return path, false
	}
	return path, s.IsDir()
}

// checkHelmVersion will return an error if the helm version is not V3
func (p *HelmChartInflationGeneratorPlugin) checkHelmVersion() error {
	stdout, err := p.runHelmCommand([]string{"version", "-c", "--short"})
	if err != nil {
		return err
	}
	r, err := regexp.Compile(`v?\d+(\.\d+)+`)
	if err != nil {
		return err
	}
	v := r.FindString(string(stdout))
	if v == "" {
		return fmt.Errorf("cannot find version string in %s", string(stdout))
	}
	if v[0] == 'v' {
		v = v[1:]
	}
	majorVersion := strings.Split(v, ".")[0]
	if majorVersion != "3" {
		return fmt.Errorf("this HelmChartInflationGeneratorPlugin requires helm V3 but got v%s", v)
	}
	return nil
}

func NewResMapFactory() *resmap.Factory {
	resourceFactory := resource.NewFactory(&hasher.Hasher{})
	resourceFactory.IncludeLocalConfigs = true
	return resmap.NewFactory(resourceFactory)
}

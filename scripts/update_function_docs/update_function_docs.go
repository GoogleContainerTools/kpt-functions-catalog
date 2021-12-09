// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Usage: update_function_docs <RELEASE_BRANCH>
//
// e.g. update_function_docs origin/apply-setters/v0.2
//
// The command will checkout the release branch and update the function/example
// docs with the latest patch version for the release. If the docs are updated
// then a commit is created with the changes. The manual steps left to the user
// are to push the commit to a branch and create a pull request.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"
)

func exitWithErr(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}

func runCmd(name string, arg ...string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(name, arg...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	fmt.Printf("%s\n", cmd.String())
	err := cmd.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("%s\n%s", stderr.String(), err)
	}
	return stdout.String(), err
}

func isCleanRepo() bool {
	_, err := runCmd("git", "diff-index", "--quiet", "HEAD", "--")
	if err != nil {
		return false
	}
	return true
}

func gitFetch() error {
	_, err := runCmd("git", "fetch", "--tags")
	return err
}

func gitCheckout(branch string) error {
	_, err := runCmd("git", "checkout", branch)
	return err
}

func gitAdd() error {
	_, err := runCmd("git", "add", "-u")
	return err
}

func gitCommit(msg string) error {
	formattedMsg := fmt.Sprintf("\"%s\"", msg)
	stdout, err := runCmd("git", "commit", "-m", formattedMsg)
	fmt.Printf("%v\n", stdout)
	return err
}

func gitShow() error {
	stdout, err := runCmd("git", "show")
	fmt.Printf("%v\n", stdout)
	return err
}

var (
	releaseBranchPattern = regexp.MustCompile(`[-\w]*\/(v\d*\.\d*)`)
	releaseTagPattern    = regexp.MustCompile(`.*(go|ts)\/[-\w]*\/(v\d*\.\d*\.\d*)`)
)

func dirExists(path string) bool {
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return true
	}
	return false
}

type functionRelease struct {
	FunctionName       string
	MinorVersion       string
	Language           string
	LatestPatchVersion string
	FunctionPath       string
	ExamplePaths       []string
}

// newFunctionRelease allocates and initializes a functionRelease
func newFunctionRelease(branch string) (*functionRelease, error) {
	fr := &functionRelease{}
	if !releaseBranchPattern.MatchString(branch) {
		return nil, fmt.Errorf("invalid branch format")
	}
	segments := strings.Split(branch, "/")
	fr.MinorVersion = segments[len(segments)-1]
	fr.FunctionName = segments[len(segments)-2]
	if err := fr.readLatestPatchVersion(); err != nil {
		return nil, err
	}
	if err := fr.readDocPaths(); err != nil {
		return nil, err
	}
	return fr, nil
}

// readLatestPatchVersion of the release from git tags
func (fr *functionRelease) readLatestPatchVersion() error {
	if fr.FunctionName == "" || fr.MinorVersion == "" {
		return fmt.Errorf("missing function name and/or minor version")
	}
	tags, err := runCmd("git", "tag")
	if err != nil {
		return err
	}
	funcPattern := fmt.Sprintf("%s/%s", fr.FunctionName, fr.MinorVersion)
	var lang, latestPatchVersion string
	for _, tag := range strings.Split(tags, "\n") {
		if !strings.Contains(tag, funcPattern) || !releaseTagPattern.MatchString(tag) {
			continue
		}
		segments := strings.Split(tag, "/")
		patchVersion := segments[len(segments)-1]
		if latestPatchVersion == "" ||
			semver.Compare(patchVersion, latestPatchVersion) == 1 {
			latestPatchVersion = patchVersion
			lang = segments[len(segments)-3]
		}
	}
	if latestPatchVersion == "" || lang == "" {
		return fmt.Errorf("could not find matching tag for release branch")
	}
	fr.Language = lang
	fr.LatestPatchVersion = latestPatchVersion
	return nil
}

// readDocPaths and set FunctionPath and ExamplePaths
func (fr *functionRelease) readDocPaths() error {
	executablePath, err := os.Executable()
	if err != nil {
		return err
	}
	repoBase := filepath.Dir(filepath.Dir(filepath.Dir(executablePath)))
	fr.FunctionPath = filepath.Join(repoBase, "functions", fr.Language, fr.FunctionName)
	examplesPath := filepath.Join(repoBase, "examples")

	if !dirExists(fr.FunctionPath) {
		fr.FunctionPath = filepath.Join(repoBase, "contrib", "functions", fr.Language, fr.FunctionName)
		examplesPath = filepath.Join(repoBase, "contrib", "examples")
		if !dirExists(fr.FunctionPath) {
			return fmt.Errorf("function doc paths not found %s", fr.FunctionPath)
		}
	}
	if err = fr.parseMetadata(examplesPath); err != nil {
		return err
	}
	return nil
}

// parseMetadata from metadata.yaml and set ExamplePaths
func (fr *functionRelease) parseMetadata(examplesPath string) error {
	type metadata struct {
		ExamplePackageUrls []string `yaml:"examplePackageURLs"`
	}
	if fr.FunctionPath == "" {
		return fmt.Errorf("expected FunctionPath in parseMetadata")
	}

	metadataPath := filepath.Join(fr.FunctionPath, "metadata.yaml")
	var md metadata
	yamlFile, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, &md)
	if err != nil {
		return err
	}
	for _, exampleURL := range md.ExamplePackageUrls {
		segments := strings.Split(exampleURL, "/")
		exampleName := segments[len(segments)-1]
		examplePath := filepath.Join(examplesPath, exampleName)
		if !dirExists(examplePath) {
			return fmt.Errorf("example dir does not exist: %s", examplePath)
		}
		fr.ExamplePaths = append(fr.ExamplePaths, examplePath)
	}
	return nil
}

// updateDoc with the latest version in place at specified path
func (fr *functionRelease) updateDoc(path string) error {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	versionGroup := `(unstable|v\d*\.\d*\.\d*|v\d*\.\d*)`
	// replace tags with patch e.g. apply-setters:v1.0.1, apply-setters/v1.0.1
	tagPattern := regexp.MustCompile(
		fmt.Sprintf(`%s(:|/)%s`, fr.FunctionName, versionGroup))
	contents = tagPattern.ReplaceAll(contents,
		[]byte(fmt.Sprintf(`%s${1}%s`, fr.FunctionName, fr.LatestPatchVersion)))
	// replace url with minor e.g. https://catalog.kpt.dev/apply-setters/v1.0
	urlPattern := regexp.MustCompile(
		fmt.Sprintf(`https://catalog\.kpt\.dev/%s/%s`, fr.FunctionName, versionGroup))
	contents = urlPattern.ReplaceAll(contents,
		[]byte(fmt.Sprintf(`https://catalog.kpt.dev/%s/%s`, fr.FunctionName, fr.MinorVersion)))

	if err = os.WriteFile(path, contents, 0644); err != nil {
		return err
	}
	return nil
}

// updateDocs calls updateDoc for each function doc
func (fr *functionRelease) updateDocs() error {
	functionReadme := filepath.Join(fr.FunctionPath, "README.md")
	if err := fr.updateDoc(functionReadme); err != nil {
		return err
	}
	for _, example := range fr.ExamplePaths {
		exampleReadme := filepath.Join(example, "README.md")
		if err := fr.updateDoc(exampleReadme); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var err error
	if len(os.Args) < 2 {
		exitWithErr(fmt.Errorf("usage: update_function_docs <RELEASE_BRANCH>"))
	}
	releaseBranch := os.Args[1]
	if !isCleanRepo() {
		exitWithErr(fmt.Errorf("dirty repo"))
	}
	if err = gitFetch(); err != nil {
		exitWithErr(err)
	}
	if err = gitCheckout(releaseBranch); err != nil {
		exitWithErr(err)
	}
	fr, err := newFunctionRelease(releaseBranch)
	if err != nil {
		exitWithErr(err)
	}
	if err = fr.updateDocs(); err != nil {
		exitWithErr(err)
	}
	if isCleanRepo() {
		exitWithErr(fmt.Errorf("docs up to date"))
	}
	if err = gitAdd(); err != nil {
		exitWithErr(err)
	}
	msg := fmt.Sprintf("docs: Update tags for %s/%s/%s",
		fr.Language, fr.FunctionName, fr.LatestPatchVersion)
	if err = gitCommit(msg); err != nil {
		exitWithErr(err)
	}
	if err = gitShow(); err != nil {
		exitWithErr(err)
	}
}

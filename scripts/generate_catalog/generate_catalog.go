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
// Usage: generate_catalog SRC_REPO_DIR/ DEST_MD_DIR/
//
// The command will create a README.md file under DEST_MD_DIR/ containing a table
// of collected functions. Source files for the catalog will also appear in
// DEST_MD_DIR/
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: generate-catalog SRC_REPO_DIR/ DEST_MD_DIR/\n")
		os.Exit(1)
	}
	source, err := filepath.Abs(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	dest, err := filepath.Abs(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	branches, err := getBranches()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	functions := getFunctions(branches, source, dest)

	err = writeFunctionIndex(functions, source, dest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	err = writeExampleIndex(functions, source, dest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

type function struct {
	FunctionName      string
	VersionToExamples map[string]map[string]example
	LatestVersion     string
	Path              string
	Description       string
	Tags              string
}

type example struct {
	LocalExamplePath  string
	RemoteExamplePath string
	RemoteSourcePath  string
}

type metadata struct {
	Image              string
	Description        string
	Tags               []string
	SourceUrl          string   `yaml:"sourceURL"`
	ExamplePackageUrls []string `yaml:"examplePackageURLs"`
}

var (
	// Match start of a version such as v1.9.1
	branchSemverPrefix = regexp.MustCompile(`[-\w]*\/(v\d*\.\d*)`)
	functionDirPrefix  = regexp.MustCompile(`.+/functions/`)
	exampleDirPrefix   = regexp.MustCompile(`.+/examples/`)
)

func getBranches() ([]string, error) {
	verBranches := make([]string, 0)

	var buf bytes.Buffer
	cmd := exec.Command("git", "branch", "-a")
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return verBranches, err
	}

	for _, branch := range strings.Split(buf.String(), "\n") {
		if branchSemverPrefix.MatchString(branch) {
			verBranches = append(verBranches, strings.TrimSpace(branch))
		}
	}
	return verBranches, err
}

func getFunctions(branches []string, source string, dest string) []function {
	functions := make(map[string]function)
	for _, b := range branches {
		segments := strings.Split(b, "/")
		funcName := segments[len(segments)-2]
		minorVersion := segments[len(segments)-1]
		funcDest := filepath.Join(dest, funcName)
		versionDest := filepath.Join(funcDest, minorVersion)
		relativeFuncPath, err := getRelativeFunctionPath(source, funcName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		err = copyExamples(b, funcName, funcDest, versionDest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		err = copyReadme(b, funcName, relativeFuncPath, versionDest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		f := functions[funcName]
		f.FunctionName = funcName
		if f.VersionToExamples == nil {
			f.VersionToExamples = make(map[string]map[string]example)
		}
		metadataPath := strings.TrimSpace(fmt.Sprintf("%v:%v", b, filepath.Join(relativeFuncPath, "metadata.yaml")))
		f, err = parseMetadata(f, metadataPath, minorVersion, versionDest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		functions[funcName] = f
	}

	flattenedFunctions := make([]function, 0)
	for _, f := range functions {
		flattenedFunctions = append(flattenedFunctions, f)
	}
	sort.Slice(flattenedFunctions, func(i int, j int) bool {
		return flattenedFunctions[i].FunctionName < flattenedFunctions[j].FunctionName
	})
	return flattenedFunctions
}

func copyExamples(b string, funcName string, funcDest string, versionDest string) error {
	exampleSource := fmt.Sprintf("examples/%v", funcName)

	// Prepare destination for versioned examples.
	err := os.MkdirAll(funcDest, 0744)
	if err != nil {
		return err
	}

	// Copy examples for the function's version to a temporary directory.
	tempDir, err := ioutil.TempDir("", "examples")
	if err != nil {
		return err
	}
	cmd := exec.Command("git", fmt.Sprintf("--work-tree=%v", tempDir), "checkout", b, "--", exampleSource)
	err = cmd.Run()
	if err != nil {
		return err
	}

	// Move example content to the site's example directory.
	err = os.Rename(filepath.Join(tempDir, "examples", funcName), versionDest)
	if err != nil {
		return err
	}

	return nil
}

func copyReadme(b string, funcName string, relativeFuncPath string, versionDest string) error {
	// Copy README for the function's version to the example directory.
	tempDir, err := ioutil.TempDir("", "functions")
	if err != nil {
		return err
	}
	cmd := exec.Command("git", fmt.Sprintf("--work-tree=%v", tempDir), "checkout", b, "--", filepath.Join(relativeFuncPath, "README.md"))
	err = cmd.Run()
	if err != nil {
		return err
	}

	// Find the README in the example directory.
	m, err := filepath.Glob(filepath.Join(tempDir, "functions", "*", funcName, "README.md"))
	if err != nil {
		return err
	}

	// Move the README to the destination directory.
	err = os.Rename(m[0], filepath.Join(versionDest, "README.md"))
	if err != nil {
		return err
	}

	return nil
}

func parseMetadata(f function, metadataPath string, version string, versionDest string) (function, error) {
	var buf bytes.Buffer
	// Get the content of metadata.yaml from the appropriate release branch.
	cmd := exec.Command("git", "cat-file", "blob", metadataPath)
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return f, err
	}

	var md metadata
	yaml.Unmarshal(buf.Bytes(), &md)

	// Add examples to version map.
	if f.VersionToExamples[version] == nil {
		f.VersionToExamples[version] = make(map[string]example)
	}
	e := f.VersionToExamples[version]
	for _, exUrl := range md.ExamplePackageUrls {
		exUrlSegments := strings.Split(exUrl, "/")
		exName := exUrlSegments[len(exUrlSegments)-1]
		ex := example{
			RemoteSourcePath:  md.SourceUrl,
			RemoteExamplePath: exUrl,
			LocalExamplePath:  filepath.Join(versionDest, exName),
		}
		e[exName] = ex
	}
	f.VersionToExamples[version] = e
	if semver.Compare(f.LatestVersion, version) == 1 {
		return f, nil
	}

	// If this is the latest version,
	// update latest version, default path, description and rags.
	f.LatestVersion = version
	f.Path = versionDest
	f.Description = md.Description
	sort.Sort(sort.StringSlice(md.Tags))
	f.Tags = strings.Join(md.Tags, ",")

	return f, nil
}

func getRelativeFunctionPath(source string, funcName string) (string, error) {
	// Find the directory for the function's source.
	m, err := filepath.Glob(filepath.Join(source, "functions", "*", funcName))
	if err != nil {
		return "", err
	}

	return functionDirPrefix.ReplaceAllString(m[0], "functions/"), nil
}

func writeFunctionIndex(functions []function, source string, dest string) error {
	out := []string{"# KPT Function Catalog", "", "| Name | Description | Tags |", "| ---- | ----------- | ---- |"}
	for _, f := range functions {
		functionEntry := fmt.Sprintf("| [%v](%v/) | %v | %v |", f.FunctionName, strings.Replace(f.Path, filepath.Join(source, "examples"), "", 1), f.Description, f.Tags)
		out = append(out, functionEntry)
	}

	o := strings.Join(out, "\n")
	err := ioutil.WriteFile(filepath.Join(dest, "README.md"), []byte(o), 0744)
	return err
}

func writeExampleIndex(functions []function, source string, dest string) error {
	// Key a function's version's examples by the function's name -> version
	functionVersionMap := make(map[string]map[string]map[string]example)
	for _, f := range functions {
		vToE := make(map[string]map[string]example)
		for v, examples := range f.VersionToExamples {
			exampleToPaths := make(map[string]example)
			for exName, ex := range examples {
				e := ex
				e.LocalExamplePath = strings.Replace(ex.LocalExamplePath, filepath.Join(source, "examples"), "", 1)
				exampleToPaths[exName] = e
			}
			vToE[v] = exampleToPaths
		}
		functionVersionMap[f.FunctionName] = vToE
	}

	funcJson, err := json.Marshal(functionVersionMap)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(dest, "catalog.json"), funcJson, 0600)
	return err
}

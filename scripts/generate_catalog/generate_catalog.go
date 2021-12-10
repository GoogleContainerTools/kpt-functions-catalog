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

var (
	branchesToSkip = []string{
		"sops/v0.2",
	}
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
	curatedFns, contribFns, err := writeFunctionIndex(functions, source, dest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	err = writeExampleIndex(curatedFns, source, dest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	err = writeExampleIndex(contribFns, source, filepath.Join(dest, "contrib"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

type function struct {
	FunctionName      string
	ImagePath         string
	VersionToExamples map[string]map[string]example
	LatestVersion     string
	Path              string
	Description       string
	Tags              string
	Gcp               bool
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
	Hidden             bool
}

var (
	// Match start of a version such as v1.9.1
	branchSemverPrefix = regexp.MustCompile(`[-\w]*\/(v\d*\.\d*)`)
	functionDirPrefix  = regexp.MustCompile(`.+/functions/`)
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
		skipCurrentBranch := false
		for _, branchToSkip := range branchesToSkip {
			if strings.Contains(branch, branchToSkip) {
				skipCurrentBranch = true
				break
			}
		}
		if skipCurrentBranch {
			continue
		}

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
		relativeFuncPath, err := getRelativeFunctionPath(source, funcName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}
		versionDest := filepath.Join(dest, funcName, minorVersion)
		if strings.Contains(relativeFuncPath, "contrib") {
			versionDest = filepath.Join(dest, "contrib", funcName, minorVersion)
		}

		// Functions with the hidden field enabled should not be processed.
		metadataPath := strings.TrimSpace(fmt.Sprintf("%v:%v", b, filepath.Join(relativeFuncPath, "metadata.yaml")))
		md, err := getMetadata(metadataPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting metadata for %q from %q: %v\n", funcName, b, err)
			os.Exit(1)
		}
		if md.Hidden {
			continue
		}
		err = copyExamples(b, md.ExamplePackageUrls, versionDest, minorVersion)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting examples for %q from %q: %v\n", funcName, b, err)
			os.Exit(1)
		}

		err = copyReadme(b, funcName, relativeFuncPath, versionDest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting README for %q from %q: %v\n", funcName, b, err)
			os.Exit(1)
		}

		f := functions[funcName]
		f.FunctionName = funcName
		if f.VersionToExamples == nil {
			f.VersionToExamples = make(map[string]map[string]example)
		}

		functions[funcName] = parseMetadata(f, md, minorVersion, versionDest)
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

func copyExamples(b string, exampleSources []string, versionDest, minorVersion string) error {
	// Copy examples for the function's version to a temporary directory.
	tempDir, err := ioutil.TempDir("", "examples")
	if err != nil {
		return err
	}

	// Prepare destination for versioned examples.
	err = os.MkdirAll(versionDest, 0744)
	if err != nil {
		return err
	}

	for _, exampleSource := range exampleSources {
		splitedPaths := strings.SplitN(exampleSource, minorVersion+string(filepath.Separator), 2)
		if len(splitedPaths) != 2 {
			return fmt.Errorf("expect 2 substring after spliting %q by %q", exampleSource, minorVersion+string(filepath.Separator))
		}
		relativePath := splitedPaths[1]
		// Fetch example into temporary directory.
		cmd := exec.Command("git", fmt.Sprintf("--work-tree=%v", tempDir), "checkout", b, "--", relativePath)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("Error running %v: %v", cmd, err)
		}

		exampleName := filepath.Base(relativePath)
		src := filepath.Join(tempDir, relativePath)
		dest := filepath.Join(versionDest, exampleName)

		// Move example content to the site's example directory.
		if err = os.RemoveAll(dest); err != nil {
			return err
		}
		err = os.Rename(src, dest)
		if err != nil {
			return err
		}

	}

	return nil
}

func copyReadme(b string, funcName string, relativeFuncPath string, versionDest string) error {
	// Copy README for the function's version to the function's directory.
	tempDir, err := ioutil.TempDir("", "functions")
	if err != nil {
		return err
	}
	cmd := exec.Command("git", fmt.Sprintf("--work-tree=%v", tempDir), "checkout", b, "--", filepath.Join(relativeFuncPath, "README.md"))
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Error running %v: %v", cmd, err)
	}

	// Find the README in the function's directory.
	m, err := filepath.Glob(filepath.Join(tempDir, "functions", "*", funcName, "README.md"))
	if err != nil {
		return err
	}

	if m == nil {
		m, err = filepath.Glob(filepath.Join(tempDir, "contrib", "functions", "*", funcName, "README.md"))
		if err != nil {
			return err
		}
	}

	// Move the README to the destination directory.
	err = os.Rename(m[0], filepath.Join(versionDest, "README.md"))
	if err != nil {
		return err
	}

	return nil
}

func getMetadata(metadataPath string) (metadata, error) {
	var buf bytes.Buffer
	var md metadata
	// Get the content of metadata.yaml from the appropriate release branch.
	cmd := exec.Command("git", "cat-file", "blob", metadataPath)
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return md, err
	}

	yaml.Unmarshal(buf.Bytes(), &md)
	return md, nil
}

func parseMetadata(f function, md metadata, version string, versionDest string) function {

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
		return f
	}

	// If this is the latest version,
	// update latest version, default path, description and rags.
	f.LatestVersion = version
	f.Path = versionDest
	f.Description = md.Description
	functionTags := make([]string, 0)
	for _, tag := range md.Tags {
		normalizedTag := strings.ToLower(tag)
		if normalizedTag == "gcp" {
			f.Gcp = true
		} else {
			functionTags = append(functionTags, normalizedTag)
		}
	}
	sort.Strings(functionTags)
	f.Tags = strings.Join(functionTags, ", ")
	f.ImagePath = md.Image

	return f
}

func getRelativeFunctionPath(source string, funcName string) (string, error) {
	// Find the directory for the function's source.
	sourcePattern := filepath.Join(source, "functions", "*", funcName)
	m, err := filepath.Glob(sourcePattern)
	if err != nil {
		return "", err
	}
	if m != nil {
		return functionDirPrefix.ReplaceAllString(m[0], "functions/"), nil
	}

	contribPattern := filepath.Join(source, "contrib", "functions", "*", funcName)
	m, err = filepath.Glob(contribPattern)
	if err != nil {
		return "", err
	}
	if m == nil {
		return "", fmt.Errorf("Could not find a function with the following name: %v", funcName)
	}
	return functionDirPrefix.ReplaceAllString(m[0], "contrib/functions/"), nil
}

func writeFunctionIndex(functions []function, source string, dest string) ([]function, []function, error) {
	out := []string{"# Curated Functions Catalog", ""}
	var contribOut []string

	genericFunctions := make([]function, 0)
	gcp := make([]function, 0)
	contribFunctions := make([]function, 0)
	for _, f := range functions {
		if strings.Contains(f.ImagePath, "contrib") {
			contribFunctions = append(contribFunctions, f)
		} else {
			if f.Gcp {
				gcp = append(gcp, f)
			} else {
				genericFunctions = append(genericFunctions, f)
			}
		}
	}

	out = append(out, getFunctionTable(genericFunctions, source)...)

	if len(gcp) > 0 {
		out = append(out, "", "## GCP Functions", "")
		out = append(out, getFunctionTable(gcp, source)...)
	}

	if len(contribFunctions) > 0 {
		contribOut = append(contribOut, "# Contrib Functions Catalog", "")
		contribOut = append(contribOut, getFunctionTable(contribFunctions, source)...)
	}

	o := strings.Join(out, "\n")
	if err := ioutil.WriteFile(filepath.Join(dest, "README.md"), []byte(o), 0744); err != nil {
		return nil, nil, err
	}
	co := strings.Join(contribOut, "\n")
	if err := ioutil.WriteFile(filepath.Join(dest, "contrib", "README.md"), []byte(co), 0744); err != nil {
		return nil, nil, err
	}
	return append(genericFunctions, gcp...), contribFunctions, nil
}

func getFunctionTable(functions []function, source string) []string {
	out := []string{"| Name | Description | Tags |", "| ---- | ----------- | ---- |"}
	for _, f := range functions {
		functionEntry := fmt.Sprintf("| [%v](%v/) | %v | %v |", f.FunctionName, strings.Replace(f.Path, filepath.Join(source, "site"), "", 1), f.Description, f.Tags)
		out = append(out, functionEntry)
	}
	return out
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
				e.LocalExamplePath = strings.Replace(ex.LocalExamplePath, filepath.Join(source, "site"), "", 1)
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

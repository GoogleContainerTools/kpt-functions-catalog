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
// Usage: generate_catalog SOURCE_MD_DIR/ DEST_GO_DIR/
//
// The command will create a README.md file under DEST_GO_DIR/ containing tables
// of functions separated by function type.
// <!--mdtogo:<VARIABLE_NAME>
// ..some content..
// -->
//
// <VARIABLE_NAME> must be Name, Type or Description.
package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/mod/semver"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: generate-catalog SOURCE_MD_DIR/ DEST_GO_DIR/\n")
		os.Exit(1)
	}
	source := os.Args[1]
	dest := os.Args[2]

	mutatorFunctions, err := getFunctions(filepath.Join(source, "mutators"), "Mutator")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	validatorFunctions, err := getFunctions(filepath.Join(source, "validators"), "Validator")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	functions, err := assignFunctionDescriptions(append(mutatorFunctions, validatorFunctions...))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	err = writeIndex(functions, source, dest)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

type function struct {
	FunctionName     string
	VersionToPatches map[string][]string
	Path             string
	Description      string
	Type             string
}

func getFunctions(source string, functionType string) ([]function, error) {
	functions := make([]function, 0)

	// Reads in exampleDir/
	dirs, err := os.ReadDir(source)
	if err != nil {
		return functions, err
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			functionName := dir.Name()
			functionPath := filepath.Join(source, dir.Name())
			// Reads in paths like exampleDir/helm-inflator
			paths, err := os.ReadDir(functionPath)
			if err != nil {
				return functions, err
			}

			if len(paths) > 0 {
				if pathHasRelease(paths) {
					sort.Slice(paths, func(i, j int) bool {
						// Sort directories by the ordering of the semantic versions they represent
						return semver.Compare(paths[i].Name(), paths[j].Name()) > 0
					})
					versions := keyPatchVersionsByMinor(paths)
					if versions != nil {
						functions = append(functions,
							function{
								FunctionName:     functionName,
								VersionToPatches: versions,
								Path:             filepath.Join(functionPath, paths[0].Name()),
								Type:             functionType,
							},
						)
					}
				}
			}
		}
	}
	return functions, nil
}

var (
	// Match start of a version such as v1.9.1
	semverPrefix = regexp.MustCompile(`v\d`)
)

func pathHasRelease(paths []fs.DirEntry) bool {
	for _, path := range paths {
		if semverPrefix.FindStringSubmatch(path.Name()) != nil {
			return true
		}
	}
	return false
}

func keyPatchVersionsByMinor(paths []fs.DirEntry) map[string][]string {
	m := make(map[string][]string)
	for _, path := range paths {
		if semver.IsValid(path.Name()) {
			m[semver.MajorMinor(path.Name())] = append(m[semver.MajorMinor(path.Name())], path.Name())
		}
	}
	return m
}

func assignFunctionDescriptions(functions []function) ([]function, error) {
	for i := range functions {

		minorVersions := make([]string, 0)
		for minorVersion := range functions[i].VersionToPatches {
			minorVersions = append(minorVersions, minorVersion)
		}

		// Sort minor versions in descending order
		sort.Slice(minorVersions, func(i, j int) bool {
			return semver.Compare(minorVersions[i], minorVersions[j]) > 0
		})

		for versionIndex, version := range minorVersions {
			imagePath := fmt.Sprintf("--image=gcr.io/kpt-fn/%v:%v", functions[i].FunctionName, version)

			// Use `kpt fn doc` to obtain documentation on the function
			var buf bytes.Buffer
			cmd := exec.Command("kpt", "fn", "doc", imagePath)
			cmd.Stdout = &buf
			err := cmd.Run()
			if err != nil {
				return functions, err
			}

			// The first line of the most recent documentation is the description of the function
			firstLine, err := buf.ReadString('\n')
			if err != nil {
				return functions, err
			}

			if versionIndex == 0 {
				functions[i].Description = firstLine
			}

			// Write the entire documentation output to the appropriate version's directory.
			for _, patchVersion := range functions[i].VersionToPatches[version] {
				versionDocumentationPath := filepath.Join(filepath.Dir(functions[i].Path), patchVersion, "README.md")
				err := ioutil.WriteFile(versionDocumentationPath, []byte(firstLine+buf.String()), 0600)
				if err != nil {
					return functions, err
				}
			}
		}
	}
	return functions, nil
}

func writeIndex(functions []function, source string, dest string) error {
	mutators := []string{"## Mutators", "", "| Name | Description |", "| ---- | ----------- |"}
	validators := []string{"## Validators", "", "| Name | Description |", "| ---- | ----------- |"}
	for _, f := range functions {
		functionEntry := fmt.Sprintf("| [%v](%v/) | %v |", f.FunctionName, strings.Replace(f.Path, source, "/", 1), f.Description)

		switch f.Type {
		case "Mutator":
			mutators = append(mutators, functionEntry)
		case "Validator":
			validators = append(validators, functionEntry)

		}
	}

	out := append([]string{"# KPT Function Catalog", ""}, mutators...)
	out = append(out, "")
	out = append(out, validators...)

	o := strings.Join(out, "\n")
	err := ioutil.WriteFile(filepath.Join(dest, "README.md"), []byte(o), 0600)
	return err
}

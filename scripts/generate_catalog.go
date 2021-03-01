// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

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
	"fmt"
	"io/ioutil"
	"os"
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

	functions, err := getFunctions(source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	functions, err = parseFuncions(functions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	err = write(functions, source, dest)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

type function struct {
	FunctionName string
	Version      string
	Path         string
	Desciption   string
	Type         string
}

func getFunctions(source string) ([]function, error) {
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

			sort.Slice(paths, func(i, j int) bool {
				firstVersion := semver.Canonical(strings.ReplaceAll(paths[i].Name(), "_", "."))
				secondVersion := semver.Canonical(strings.ReplaceAll(paths[j].Name(), "_", "."))
				// Sort directories by the ordering of the semantic versions they represent
				return semver.Compare(firstVersion, secondVersion) > 0
			})

			if len(paths) > 0 {
				potentialFunctionVersion := semver.Canonical(strings.ReplaceAll(paths[0].Name(), "_", "."))
				if semver.IsValid(potentialFunctionVersion) {
					functions = append(functions,
						function{
							FunctionName: functionName,
							Version:      potentialFunctionVersion,
							Path:         filepath.Join(functionPath, paths[0].Name()),
						},
					)
				}
			}
		}
	}

	return functions, nil
}

var (
	// Capture content within a tag like <!--catalog:[Name|Description|Type] (Content)-->
	tags = regexp.MustCompile(`<!--catalog:(Name|Description|Type)\s+?([\s\S]*?)-->`)
)

func parseFuncions(functions []function) ([]function, error) {
	for i := range functions {
		b, err := ioutil.ReadFile(filepath.Join(functions[i].Path, "README.md"))
		if err != nil {
			return functions, err
		}

		markdown := string(b)
		matches := tags.FindAllStringSubmatch(markdown, 3)

		for _, match := range matches {
			switch match[1] {
			case "Name":
				functions[i].FunctionName = strings.TrimSpace(match[2])
			case "Description":
				functions[i].Desciption = strings.TrimSpace(match[2])
			case "Type":
				functions[i].Type = strings.Title(strings.TrimSpace(match[2]))
			}
		}

	}
	return functions, nil
}

func write(functions []function, source string, dest string) error {
	mutators := []string{"## Mutators", "", "| Name | Description |", "| ---- | ----------- |"}
	validators := []string{"## Validators", "", "| Name | Description |", "| ---- | ----------- |"}
	for _, f := range functions {
		functionEntry := fmt.Sprintf("| [%v](%v/) | %v |", f.FunctionName, strings.Replace(f.Path, source, "/", 1), f.Desciption)

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

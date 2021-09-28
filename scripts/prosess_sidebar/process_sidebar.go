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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: process_sidebar DEST_DIR/\n")
		os.Exit(1)
	}
	dest, err := filepath.Abs(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Download the sidebar.md from kpt repo which is the source of truth.
	res, err := http.Get("https://raw.githubusercontent.com/GoogleContainerTools/kpt/main/site/sidebar.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to get the sidebar.md file from kpt repo: %v\n", err)
		os.Exit(1)
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		fmt.Fprintf(os.Stderr, "response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read the response body: %v\n", err)
		os.Exit(1)
	}

	newSidebar, err := processSidebar(body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read the response body: %v\n", err)
		os.Exit(1)
	}

	err = ioutil.WriteFile(filepath.Join(dest, "site/sidebar.md"), newSidebar, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read the response body: %v\n", err)
		os.Exit(1)
	}
}

const (
	// To match lines like `  - [CLI](reference/cli/)`
	sidebarLineRegExp = `(\s*- \[.+\]\()(.+)(\))`
	// To match lines like `  - [Curated](https://catalog.kpt.dev/ ":target=_self")`
	sidebarLineWithTagRegExp = `(\s*- \[.+\]\()(.+?)( +.*\))`
	crossoriginAndTargetTags = ` ":crossorgin :target=_self"`
)

func processSidebar(content []byte) ([]byte, error) {
	lines := bytes.Split(content, []byte("\n"))
	// We are trying to match something like
	// - [1.1 System Requirements](book/01-getting-started/01-system-requirements.md)
	re := regexp.MustCompile(sidebarLineRegExp)
	reWithTag := regexp.MustCompile(sidebarLineWithTagRegExp)
	var newLines [][]byte
	for _, line := range lines {
		submatches := reWithTag.FindSubmatch(line)
		if submatches == nil {
			submatches = re.FindSubmatch(line)
			if submatches == nil {
				newLines = append(newLines, line)
				continue
			}
		}

		if len(submatches) != 4 {
			return nil, fmt.Errorf("%s is expected to match either %s or %s", line, sidebarLineWithTagRegExp, sidebarLineRegExp)
		}

		link := string(submatches[2])
		if !strings.Contains(link, "http") {
			u := url.URL{
				Scheme: "https",
				Host:   "kpt.dev",
				Path:   link,
			}
			newLines = append(newLines, []byte(string(submatches[1])+u.String()+crossoriginAndTargetTags+string(submatches[3])))
		} else {
			u, err := url.Parse(link)
			if err != nil {
				return nil, fmt.Errorf("unable to parse URL %s: %v", link, err)
			}
			newLines = append(newLines, []byte(string(submatches[1])+u.Path+")"))
		}
	}
	return bytes.Join(newLines, []byte("\n")), nil
}

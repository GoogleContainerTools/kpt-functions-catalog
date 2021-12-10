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
	"fmt"
	"os"
)

func exitWithErr(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
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

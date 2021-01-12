package e2etest

import (
	"fmt"
	"testing"

	"github.com/GoogleContainerTools/kpt-functions-catalog/tests/e2etest/internal/runner"
)

// TestE2E accepts a path and scans the path to find all available packages that can
// be tested. A package which contains a directory '.expected' is considered testable.
//
// The kpt package should contain a declarative function that will be tested.
//
// This driver expects a directory '.expected' in the top level of the package and
// '.expected' should contain 3 files:
//  - 'exitcode.txt' contains a single number which is expected exit code of command
//    'kpt fn run'. If this file is missed, driver will assume the expected exit code
//    is 0.
//  - 'diff.patch' is the expected diff output between original package files and
//    files after function running. The diff will be compared only when the exit code
//    matches expected and is zero.
//  - 'results.yaml' is the expected results output from the command 'kpt fn run'.
//    The results will be compared only when the exit code matches expected and is not
//    zero.
//  - 'network.txt' contains a string which indicates whether network should be enabled
//    for this test. If this file existis and the content in it is 'true' then the
//    network is accessible. Otherwise the function cannot access network.
//
// Given a package's name is 'my-pkg', this driver will copy the package into a temporary
// directory and then run command 'kpt fn run my-pkg --results-dir results'. The test
// will pass when the diff output, results output and exit code are all matched with
// expected.
//
// Git is required to generate diff output.
func TestE2E(t *testing.T) {
	err := runTests("../../..")
	if err != nil {
		t.Fatal(err)
	}
}

// runTests will scan test cases in 'path' and run all the
// tests in it. It returns an error if any of the tests fails.
func runTests(path string) error {
	cases, err := runner.ScanTestCases(path)
	if err != nil {
		return fmt.Errorf("failed to scan test cases: %w", err)
	}
	fmt.Printf("Found %d tests in %s:\n", len(*cases), path)
	for _, c := range *cases {
		fmt.Printf(" - %s\n", c)
	}
	fmt.Println("\nStart running...")
	var retErr []chan error
	for i, c := range *cases {
		retErr = append(retErr, make(chan error))
		r, err := runner.NewRunner(c)
		if err != nil {
			return fmt.Errorf("failed to run test: %w", err)
		}
		go r.Run(retErr[i])
	}
	hasError := false
	for i := range retErr {
		err := <-retErr[i]
		if err != nil {
			fmt.Printf("FAIL: %s: %s\n", (*cases)[i], err)
			hasError = true
		} else {
			fmt.Printf("PASS: %s\n", (*cases)[i])
		}
	}
	if hasError {
		return fmt.Errorf("Test failed")
	}
	return nil
}

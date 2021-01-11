package runner

import (
	"fmt"
	"os"
	"path/filepath"
)

// TestCase contains the information needed to run a test. Each test case
// run by this driver is described by a `TestCase`.
type TestCase string

// TestCases contains a list of TestCase.
type TestCases []TestCase

func isTestCase(path string, info os.FileInfo) bool {
	if !info.IsDir() {
		return false
	}

	expectedPath := filepath.Join(path, expectedDir)
	expectedInfo, err := os.Stat(expectedPath)
	if err != nil {
		return false
	}
	if !expectedInfo.IsDir() {
		return false
	}
	return true
}

// ScanTestCases will recursively scan the directory `path` and return
// a list of TestConfig found
func ScanTestCases(path string) (*TestCases, error) {
	var cases TestCases
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !isTestCase(path, info) {
			return nil
		}

		cases = append(cases, TestCase(path))

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan test cases in %s", path)
	}
	return &cases, nil
}

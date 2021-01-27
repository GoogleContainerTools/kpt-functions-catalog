package runner

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Runner runs an e2e test
type Runner struct {
	pkgPath  string
	pkgName  string
	testCase TestCase
}

const (
	expectedDir         string = ".expected"
	expectedResultsFile string = "results.yaml"
	expectedDiffFile    string = "diff.patch"
	expectedConfigFile  string = "config.yaml"
)

// NewRunner returns a new runner for pkg
func NewRunner(testCase TestCase) (*Runner, error) {
	pkg := testCase.Path
	info, err := os.Stat(pkg)
	if err != nil {
		return nil, fmt.Errorf("cannot open path %s: %w", pkg, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path %s is not a directory", pkg)
	}
	return &Runner{
		pkgPath:  pkg,
		pkgName:  filepath.Base(pkg),
		testCase: testCase,
	}, nil
}

// Run runs the test.
func (r *Runner) Run() error {
	fmt.Printf("Running test against package %s\n", r.pkgName)
	tmpDir, err := ioutil.TempDir("", "kpt-fn-catalog-e2e-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	tmpPkgPath := filepath.Join(tmpDir, r.pkgName)
	// create result dir
	resultsPath := filepath.Join(tmpDir, "results")
	err = os.Mkdir(resultsPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create results dir %s: %w", resultsPath, err)
	}

	// copy package to temp directory
	err = copyDir(r.pkgPath, tmpPkgPath)
	if err != nil {
		return fmt.Errorf("failed to copy package: %w", err)
	}

	// init and commit package files
	err = r.preparePackage(tmpPkgPath)
	if err != nil {
		return fmt.Errorf("failed to prepare package: %w", err)
	}

	// run function
	// TODO: change to pipeline when it's ready
	kptArgs := []string{"fn", "run", tmpPkgPath, "--results-dir", resultsPath}
	if r.testCase.Config.Network {
		kptArgs = append(kptArgs, "--network")
	}
	var output string
	var fnErr error
	for i := 0; i < r.testCase.Config.RunTimes; i++ {
		output, fnErr = runCommand("", "kpt", kptArgs)
		if fnErr != nil {
			// if kpt fn run returns error, we should compare
			// the result
			break
		}
	}

	// run formatter
	_, err = runCommand("", "kpt", []string{"cfg", "fmt", tmpPkgPath})
	if err != nil {
		return fmt.Errorf("failed to run kpt cfg fmt: %w", err)
	}

	// compare results
	err = r.compareResult(fnErr, tmpPkgPath, resultsPath)
	if err != nil {
		return fmt.Errorf("%w\nkpt output:\n%s", err, output)
	}
	return nil
}

func (r *Runner) preparePackage(pkgPath string) error {
	err := gitInit(pkgPath)
	if err != nil {
		return err
	}

	err = gitAddAll(pkgPath)
	if err != nil {
		return err
	}

	return gitCommit(pkgPath, "first")
}

func (r *Runner) compareResult(exitErr error, tmpPkgPath, resultsPath string) error {
	expected, err := newExpected(tmpPkgPath)
	if err != nil {
		return err
	}
	// get exit code
	exitCode := 0
	if e, ok := exitErr.(*exec.ExitError); ok {
		exitCode = e.ExitCode()
	} else if exitErr != nil {
		return fmt.Errorf("cannot get exit code from %w", exitErr)
	}

	if exitCode != r.testCase.Config.ExitCode {
		return fmt.Errorf("actual exit code %d doesn't match expected %d", exitCode, r.testCase.Config.ExitCode)
	}

	if exitCode != 0 {
		actual, err := readActualResults(resultsPath)
		if err != nil {
			return fmt.Errorf("failed to read actual results: %w", err)
		}
		if actual != expected.Results {
			return fmt.Errorf("actual results doesn't match expected\nActual\n===\n%s\nExpected\n===\n%s",
				actual, expected.Results)
		}
		return nil
	}

	// compare diff
	actual, err := readActualDiff(tmpPkgPath)
	if err != nil {
		return fmt.Errorf("failed to read actual diff: %w", err)
	}
	if actual != expected.Diff {
		return fmt.Errorf("actual diff doesn't match expected\nActual\n===\n%s\nExpected\n===\n%s",
			actual, expected.Diff)
	}
	return nil
}

func readActualResults(resultsPath string) (string, error) {
	l, err := ioutil.ReadDir(resultsPath)
	if err != nil {
		return "", fmt.Errorf("failed to get files in results dir: %w", err)
	}
	if len(l) != 1 {
		return "", fmt.Errorf("unexpected results files number %d, should be 1", len(l))
	}
	resultsFile := l[0].Name()
	actualResults, err := ioutil.ReadFile(filepath.Join(resultsPath, resultsFile))
	if err != nil {
		return "", fmt.Errorf("failed to read actual results: %w", err)
	}
	return strings.TrimSpace(string(actualResults)), nil
}

func readActualDiff(path string) (string, error) {
	err := gitAddAll(path)
	if err != nil {
		return "", err
	}
	err = gitCommit(path, "second")
	if err != nil {
		return "", err
	}
	// diff with first commit
	actualDiff, err := gitDiff(path, "HEAD^", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(actualDiff)), nil
}

// expected contains the expected result for the function running
type expected struct {
	Results string
	Diff    string
}

func newExpected(path string) (expected, error) {
	e := expected{}
	// get expected results
	expectedResults, err := ioutil.ReadFile(filepath.Join(path, expectedDir, expectedResultsFile))
	if os.IsNotExist(err) {
		e.Results = ""
	} else if err != nil {
		return e, fmt.Errorf("failed to read expected results: %w", err)
	} else {
		e.Results = strings.TrimSpace(string(expectedResults))
	}

	// get expected diff
	expectedDiff, err := ioutil.ReadFile(filepath.Join(path, expectedDir, expectedDiffFile))
	if os.IsNotExist(err) {
		e.Diff = ""
	} else if err != nil {
		return e, fmt.Errorf("failed to read expected diff: %w", err)
	} else {
		e.Diff = strings.TrimSpace(string(expectedDiff))
	}

	return e, nil
}

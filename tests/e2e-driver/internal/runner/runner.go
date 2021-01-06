package runner

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Runner runs an e2e test
type Runner struct {
	pkgPath string
	pkgName string
	network bool
}

const (
	expectedDir          string = ".expected"
	expectedExitCodeFile string = "exitcode.txt"
	expectedResultsFile  string = "results.yaml"
	expectedDiffFile     string = "diff.patch"
)

// NewRunner returns a new runner for pkg
func NewRunner(pkg string, network bool) (*Runner, error) {
	info, err := os.Stat(pkg)
	if err != nil {
		return nil, fmt.Errorf("cannot open path %s: %w", pkg, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path %s is not a directory", pkg)
	}
	return &Runner{
		pkgPath: pkg,
		pkgName: filepath.Base(pkg),
		network: network,
	}, nil
}

// Run runs the test
func (r *Runner) Run() error {
	fmt.Printf("Running test against package %s\n", r.pkgName)
	tmpDir, err := ioutil.TempDir("", "kpt-fn-catalog-e2e-*")
	fmt.Printf("Working directory: %s\n", tmpDir)
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
	if r.network {
		kptArgs = append(kptArgs, "--network")
	}
	o, err := runCommand("", "kpt", kptArgs)

	// compare results
	err = r.compareResult(err, tmpPkgPath, resultsPath)
	if err != nil {
		return fmt.Errorf("%w\nkpt output:\n%s", err, o)
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
	// get exit code
	exitCode := 0
	if e, ok := exitErr.(*exec.ExitError); ok {
		exitCode = e.ExitCode()
	} else if exitErr != nil {
		return fmt.Errorf("cannot get exit code from %w", exitErr)
	}
	// compare exit code
	b, err := ioutil.ReadFile(filepath.Join(tmpPkgPath, expectedDir, expectedExitCodeFile))
	var expectedExitCode int
	if os.IsNotExist(err) {
		expectedExitCode = 0
	} else if err != nil {
		return fmt.Errorf("failed to read expected exit code: %w", err)
	} else {
		expectedExitCode, err = strconv.Atoi(strings.TrimSpace(string(b)))
		if err != nil {
			return fmt.Errorf("cannot convert exit code %s to int: %w", b, err)
		}
	}

	if exitCode != expectedExitCode {
		return fmt.Errorf("actual exit code %d doesn't match expected %d", exitCode, expectedExitCode)
	}

	if exitCode != 0 {
		// compare results
		l, err := ioutil.ReadDir(resultsPath)
		if err != nil {
			return fmt.Errorf("failed to get files in results dir: %w", err)
		}
		if len(l) != 1 {
			return fmt.Errorf("unexpected results files number %d, should be 1", len(l))
		}
		resultsFile := l[0].Name()
		actualResults, err := ioutil.ReadFile(filepath.Join(resultsPath, resultsFile))
		if err != nil {
			return fmt.Errorf("failed to read actual results: %w", err)
		}
		expectedResults, err := ioutil.ReadFile(filepath.Join(tmpPkgPath, expectedDir, expectedResultsFile))
		if err != nil {
			return fmt.Errorf("failed to read expected results: %w", err)
		}
		actual := strings.TrimSpace(string(actualResults))
		expected := strings.TrimSpace(string(expectedResults))
		if actual != expected {
			return fmt.Errorf("actual results doesn't match expected\nActual\n===\n%s\nExpected\n===\n%s",
				actual, expected)
		}
		return nil
	}

	// compare diff
	err = gitAddAll(tmpPkgPath)
	if err != nil {
		return err
	}
	err = gitCommit(tmpPkgPath, "second")
	if err != nil {
		return err
	}
	// diff with first commit
	actualDiff, err := gitDiff(tmpPkgPath, "HEAD^", "HEAD")
	if err != nil {
		return err
	}
	expectedDiff, err := ioutil.ReadFile(filepath.Join(tmpPkgPath, expectedDir, expectedDiffFile))
	if err != nil {
		return fmt.Errorf("failed to read expected diff: %w", err)
	}
	actual := strings.TrimSpace(actualDiff)
	expected := strings.TrimSpace(string(expectedDiff))
	if actual != expected {
		return fmt.Errorf("actual diff doesn't match expected\nActual\n===\n%s\nExpected\n===\n%s",
			actual, expected)
	}
	return nil
}

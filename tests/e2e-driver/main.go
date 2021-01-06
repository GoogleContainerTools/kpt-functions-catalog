package main

import (
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/tests/e2e-driver/internal/runner"
	"github.com/spf13/cobra"
)

func main() {
	cmd := cobra.Command{
		Use:   "e2e-driver",
		Short: "e2e-driver is used to run e2e tests in kpt-functions-catalog",
		Long:  usage(),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := runner.ConfigFromFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to get config test: %w", err)
			}
			for _, c := range config.Configs {
				r, err := runner.NewRunner(c.PkgPath, c.Network)
				if err != nil {
					return fmt.Errorf("failed to run test: %w", err)
				}
				err = r.Run()
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// usage returns the usage string
func usage() string {
	return `
e2e-driver accepts an argument which is a path to YAML config file. Here is an example of
the config file:

configs:
- pkgPath: my-pkg
  network: true
- pkgPath: another-pkg

'configs' field is a list of test configs. Each config can have 2 fields:
 - pkgPath: Path to the package to be tested.
 - network: Set to 'true' if the function in package needs network access. Default is
   false.

The kpt package should contain a declarative function that will be tested.

This driver expects a directory '.expected' in the top level of the package and
'.expected' should contain 3 files:
 - 'exitcode.txt' contains a single number which is expected exit code of command
   'kpt fn run'. If this file is missed, driver will assume the expected exit code
   is 0.
 - 'diff.patch' is the expected diff output between original package files and
   files after function running. The diff will be compared only when the exit code
   matches expected and is zero.
 - 'results.yaml' is the expected results output from the command 'kpt fn run'.
   The results will be compared only when the exit code matches expected and is not
   zero.

Given a package's name is 'my-pkg', this driver will copy the package into a temporary
directory and then run command 'kpt fn run my-pkg --results-dir results'. The test
will pass when the diff output, results output and exit code are all matched with
expected.

Git is required to generate diff output.
`
}

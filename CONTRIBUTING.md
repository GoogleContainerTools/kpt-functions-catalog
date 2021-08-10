# Contributing to kpt-functions-catalog

We'd love to accept your contributions to this project. There are just a few
small guidelines you need to follow.

## Contributor License Agreement

Contributions to this project must be accompanied by a Contributor License
Agreement. You (or your employer) retain the copyright to your contribution;
this simply gives us permission to use and redistribute your contributions as
part of the project. Head over to <https://cla.developers.google.com/> to see
your current agreements on file or to sign a new one.

You generally only need to submit a CLA once, so if you've already submitted
one (even if it was for a different project), you probably don't need to do it
again.

## Code reviews

All submissions, including submissions by project members, require review. We
use GitHub pull requests for this purpose. Consult [GitHub Help] for more
information on using pull requests.

## Community Guidelines

This project follows [Google's Open Source Community Guidelines] and
a [Code of Conduct].

## Style Guides

Contributions are required to follow these style guides:

- [Error Message Style Guide]
- [Documentation Style Guide]

## How to Contribute

### Repo Layout

```shell
├── examples: Home for all function examples
│     ├── contrib: Home for all contrib function examples
│     ├── curated_function_bar_example
│     └── curated_function_foo_example
├── functions
│     ├── contrib: Home for all contrib function source code
│     │     └── ts: Home for all typescript-based contrib function source code
│     │         ├── Makefile
│     │         └── contrib_ts_function_foo
│     ├── go: Home for all golang-based curated function source code
│     │     ├── Makefile
│     │     ├── curated_go_function_bar
│     │     └── curated_go_function_foo
│     └── ts: Home for all typescript-based curated function source code
│         ├── Makefile
│         ├── curated_ts_function_bar
│         └── curated_ts_function_foo
├── scripts
└── tests: Home for e2e tests
```

For each function, its files spread in the follow places:

- `functions/` directory: Each function must have its own directory in one
  of `functions/` sub-directory. In each function's directory, it must have the
  following:
    - Source code (and unit tests).
    - A README.md file serving as the usage doc and will be shown in
      the [catalog website].
        - golang-based functions should follow [this template][golang-template].
        - typescript-based functions should follow [this template][ts-template].
    - A metadata.yaml file that follows the function metadata schema.
    - A Dockerfile to build the docker container.
- `examples/` directory: It contains examples for functions, and these examples
  are also being tested as e2e tests. Each function should have at least one
  example here. There must be a README.md file in each example directory, and it
  should follow the [template][example-template].
- The `tests/` directory contains additional e2e tests.

For golang-based functions, you need to generate some doc related variables from
the `README.md` by running

```shell
$ cd functions/go
$ make generate
```

### Tests

#### Unit Tests

To run all unit tests

```shell
$ make unit-test
```

#### E2E Tests

The e2e tests are the recommended way to test functions in the catalog. They are
very easy to write and set up with our e2e test harness. You can find all the
supported options and expected test directory
structure [here][e2e test harness doc].

You can choose to put the e2e test in either the `examples/` directory or in the
`tests/` directory depending on if it is worthwhile to be shown as an example.

To test a specific example or the e2e test, run

```shell
$ cd tests/e2etest
$ go test -v ./... -run TestE2E/../../examples/$EXAMPLE_NAME
```

To update the expected `diff.patch` or `results.yaml`, run
```shell
# Update one example
$ KPT_E2E_UPDATE_EXPECTED=true go test -v ./... -run TestE2E/../../examples/$EXAMPLE_NAME

# Update all examples
$ KPT_E2E_UPDATE_EXPECTED=true go test -v ./...
```


Most contributors don't need this, but if you happen to need to test all
examples and e2e tests, run the following command

```shell
$ make e2e-test
```

#### Doc Verifier

We have a script to ensure the usage docs and the examples are consistent.
Please ensure it's passing by running:

```shell
$ ./scripts/verify-docs.py
```

This script requires Python 3, `pyyaml` and `mdrip` which is a CLI tool.

To install `pyyaml`, run the following command:

```shell
pip install pyyaml
```

To install `mdrip`, run the following commands:

```shell
$ go get github.com/russross/blackfriday/v2@v2.0.1
$ go get github.com/monopole/mdrip@v1.0.2
```

And you need to ensure `$GOPATH/bin` is in your `PATH`.

### Change Existing Functions

You must follow the layout convention when you make changes to existing
functions.

If you implement a new feature, you must add a new example or modify existing
one to cover it.

If you fix a bug, you must add (unit or e2e) tests to cover that.

### Contribute New Functions

You must follow the layout convention when you contribute new functions.

You need to add new function name to the respective language Makefiles.

- `functions/go/Makefile` for golang.
- `functions/ts/Makefile` for typescript.

## Contact Us

Do you need a review or release of functions? We’d love to hear from you!

* Message our [Slack channel]
* Join our [email list]

[Google's Open Source Community Guidelines]: https://opensource.google.com/conduct/

[Code of Conduct]: CODE_OF_CONDUCT.md

[catalog website]: https://catalog.kpt.dev/

[e2e test harness doc]: https://github.com/GoogleContainerTools/kpt/blob/main/pkg/test/runner/README.md

[golang-template]: https://raw.githubusercontent.com/GoogleContainerTools/kpt-functions-catalog/master/functions/go/_template/README.md

[ts-template]: https://raw.githubusercontent.com/GoogleContainerTools/kpt-functions-catalog/master/functions/ts/_template/README.md

[example-template]: https://raw.githubusercontent.com/GoogleContainerTools/kpt-functions-catalog/master/examples/_template/README.md

[Slack channel]: https://kubernetes.slack.com/channels/kpt/

[email list]: https://groups.google.com/forum/?oldui=1#!forum/kpt-users

[error message style guide]: https://github.com/GoogleContainerTools/kpt/blob/main/docs/style-guides/errors.md

[documentation style guide]: https://github.com/GoogleContainerTools/kpt/blob/main/docs/style-guides/docs.md

[GitHub Help]: https://help.github.com/articles/about-pull-requests/

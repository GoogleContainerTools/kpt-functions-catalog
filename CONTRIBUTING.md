# Contributing to kpt-functions-catalog

We'd love to accept your contributions to this project. There are
just a few small guidelines you need to follow.

## Community Guidelines

This project follows [Google's Open Source Community Guidelines]
and a [Code of Conduct].

## How to Contribute

This catalog contains configuration functions to fetch, display, customize,
update, validate, and apply Kubernetes configuration. Contribute functions
to this catalog by making the following changes:

1. Make the code changes implementing the function and any unit tests. The code
   belongs under [functions/] in the appropriate directory.
2. Create an example of how to run your function and put it under [examples/].
3. Test running your function example imperatively and declaratively in
   [e2e tests]. These tests are run regularly and help catch regressions.
4. Document your function in the config functions catalog on the [kpt website]
   by following the [kpt contribution guidelines].
5. Reach out to the maintainers to publish the function.

Make PRs and request reviews for these changes. You can follow the pull request
changes below:

* [function implementation PR]
* [example and e2e test PR]
* [catalog documentation PR]

### Code Changes

#### Contributor License Agreement

Contributions to this project must be accompanied by a Contributor License
Agreement. You (or your employer) retain the copyright to your contribution;
this simply gives us permission to use and redistribute your contributions as
part of the project. Head over to <https://cla.developers.google.com/> to see
your current agreements on file or to sign a new one.

You generally only need to submit a CLA once, so if you've already submitted
one (even if it was for a different project), you probably don't need to do it
again.

#### Code reviews

All submissions, including submissions by project members, require review. We
use GitHub pull requests for this purpose. Consult
[GitHub Help](https://help.github.com/articles/about-pull-requests/) for more
information on using pull requests.

### Documentation Changes

This catalog is documented on the [kpt website]. Follow the
[kpt contribution guidelines] to make docs changes.

Changes to other documentation such as examples and README files can follow the
same pull request format as code changes.

## How to Release

We have spearate a release process for each language: golang and typescript.
All functions written in the language get released together. Maintainers should
create releases through the
[Github UI](https://github.com/GoogleContainerTools/kpt-functions-catalog/releases).

The release title and the release tag version should both be of the form
`release-[lang]-functions-v[version number]`. Risky changes are encouraged to be
tagged as pre-releases to confirm production readiness.

## Contact Us

Do you need a review or release of configuration functions? We’d love to hear
from you!

* Message our [Slack channel]
* Join our [email list]

[Google's Open Source Community Guidelines]: https://opensource.google.com/conduct/
[Code of Conduct]: CODE_OF_CONDUCT.md
[kpt website]: https://googlecontainertools.github.io/kpt/guides/consumer/function/
[kpt contribution guidelines]: https://github.com/GoogleContainerTools/kpt/blob/master/CONTRIBUTING.md#adding-or-updating-catalog-functions
[functions/]: functions/
[examples/]: examples/
[e2e tests]: tests/e2e.sh
[function implementation PR]: https://github.com/GoogleContainerTools/kpt-functions-catalog/pull/61/
[example and e2e test PR]: https://github.com/GoogleContainerTools/kpt-functions-catalog/pull/71
[catalog documentation PR]: https://github.com/GoogleContainerTools/kpt/pull/785/
[Slack channel]: https://kubernetes.slack.com/channels/kpt/
[email list]: https://groups.google.com/forum/?oldui=1#!forum/kpt-users

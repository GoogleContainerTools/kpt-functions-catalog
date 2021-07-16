# Release

## Branching Strategy

We maintain a release branch for every minor version of every function.

For a specific major version, only the release branch of the latest minor
version is in active mode (bugs will be fixed and documentations will be
updated), while the release branches for the other minor versions are in
maintenance mode (only critical bugs will be fixed).

## Release Process

### Release Major or Minor Versions

1. Check the CI badge status on top of the repo's README page. If the CI is
   failing on the master, we must fix it before doing a release.
1. Create a new release branch from the `master` branch. The new release branch
   name should be `<function-name>/v<major>.<minor>`. For example, if you want
   to release version `v1.2.0` for the `gatekeeper` function, you need to create
   branch `gatekeeper/v1.2`.
1. Create a PR to update the usage doc and examples
   like [this](https://github.com/GoogleContainerTools/kpt-functions-catalog/pull/476/files)
   . Please ensure the `verify-ci` check is passing. Merge the PR.
1. Go to the [releases pages] in your browser and create a new release for a
   function. The tag version format must
   be `<language>/<function-name>/<semver>`. e.g.
   `go/set-namespace/v1.2.0` and `ts/kubeval/v2.3.0`. The target branch must be
   the release branch instead of `master`. The release title should be
   `<funtion-name>: <semver>`. The breaking changes must be clearly stated in
   the release notes.
1. Send an announcement email in the [kpt users google group].

### Release Patch Versions

#### Pre-v1

For pre-v1 patch releases (e.g. `v0.1.2`), non-breaking changes (including new
features and bug fixes) are allowed.

1. Check the CI badge status on top of the repo's README page. If the CI is
   failing on the master, we must fix it before doing a release.
1. Checkout the release branch locally.
1. Merge the HEAD of the `master` branch into the target release branch.
1. If needed, update the usage doc and examples
   like [this](https://github.com/GoogleContainerTools/kpt-functions-catalog/pull/476/files)
1. Create a PR. Please ensure the `verify-ci` check is passing. Please ensure
   you merge the PR (neither `squash and merge` nor `rebase and merge`) so that
   the commits from the `master` branch are ported as is to the release branch.
1. Go to the [releases pages] in your browser and create a new release for a
   function. The tag version format must
   be `<language>/<function-name>/<semver>`. e.g.
   `go/set-namespace/v0.1.2` and `ts/kubeval/v0.2.3`. The target branch must be
   the release branch instead of `master`. The release title should be
   `<funtion-name>: <semver>`.

### Post-v1

For post-v1 patches releases (e.g. v1.2.3), only bug fixes are allowed.

1. Check the CI badge status on top of the repo's README page. If the CI is
   failing on the master, we must fix it before doing a release.
1. Ensure the desired bug fix PR(s) has been cherry-picked into the desired
   release branch.
1. If needed, update the usage doc and examples
   like [this](https://github.com/GoogleContainerTools/kpt-functions-catalog/pull/476/files)
   in a PR. Please ensure the `verify-ci` check is passing. Merge the PR with
   any merge option.
1. Go to the [releases pages] in your browser and create a new release for a
   function. The tag version format must
   be `<language>/<function-name>/<semver>`. e.g.
   `go/set-namespace/v1.2.3` and `ts/kubeval/v2.3.4`. The target branch must be
   the release branch instead of `master`. The release title should be
   `<funtion-name>: <semver>`.

[releases pages]: https://github.com/GoogleContainerTools/kpt-functions-catalog/releases

[kpt users google group]: https://groups.google.com/g/kpt-users

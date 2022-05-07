# Release Process

This doc covers the release process for the functions in the
kpt-functions-catalog repo.

1. Checking the CI badge status on top of the repo's README page. If the CI is
   failing on the master, we need to fix it before doing a release.
1. Go to the [releases pages] in your browser.
1. Click `Draft a new release` to create a new release for a function. The tag
   version format should be `{language}/{function-name}/{semver}`. e.g.
   `go/set-namespace/v1.2.3` and `ts/kubeval/v2.3.4`. The release name should be
   `{funtion-name} {semver}`. The release notes for this function should be in
   the body. 
1. Click `Publish release` button.
1. Send an announcement email in the [kpt users google group].

## Updating function docs

After creating a release, the docs for the function should be updated to reflect
the latest patch version. A script has been created to automate this process.
The `RELEASE_BRANCH` branch should already exist in the [repo] and a tag should
be created on the [releases pages]. `RELEASE_BRANCH` is in the form of
`${FUNCTION_NAME}/v${MAJOR_VERSION}.${MINOR_VERSION}`.
For example `set-namespace/v0.3`, `kubeval/v0.1`, etc.


1. Run the script:
```shell
# Fetch from upstream (assuming upstream is set to official repo)
git fetch upstream
# Make sure local release branch is up to date
# e.g. git checkout set-namespace/v0.3 && git reset --hard upstream/set-namespace/v0.3
git checkout <RELEASE_BRANCH> && git reset --hard upstream/<RELEASE_BRANCH>
# Check out latest version of the make target from master
git checkout upstream/master
# Run the make target
# e.g. RELEASE_BRANCH=set-namespace/v0.3 make update-function-docs
RELEASE_BRANCH=<RELEASE_BRANCH> make update-function-docs
```
1. The script will generate a new commit in your local repository which updates
the docs for the provided function release. Push this commit to your remote.
```shell
# Push the commit to your remote (branch name can be anything)
# e.g. git push origin HEAD:update-docs-set-namespace
git push origin HEAD:<remote-branch>
```
1. Create a pull request targeted at the release branch.

[repo]: https://github.com/GoogleContainerTools/kpt-functions-catalog
[releases pages]: https://github.com/GoogleContainerTools/kpt-functions-catalog/releases
[kpt users google group]: https://groups.google.com/g/kpt-users

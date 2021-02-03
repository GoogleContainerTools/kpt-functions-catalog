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

[releases pages]: https://github.com/GoogleContainerTools/kpt-functions-catalog/releases
[kpt users google group]: https://groups.google.com/g/kpt-users

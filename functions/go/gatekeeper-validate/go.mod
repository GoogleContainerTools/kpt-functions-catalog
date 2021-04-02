module github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gatekeeper-validate

go 1.15

require (
	github.com/open-policy-agent/frameworks/constraint v0.0.0-20201020161305-2e11d4556af8
	github.com/open-policy-agent/gatekeeper v0.0.0-20210128025445-201a78d6096e // This is v3.3.0. It has a semver major version of 2 or higher and is not a Go module yet.
	k8s.io/apimachinery v0.18.6
	sigs.k8s.io/kustomize/kyaml v0.10.13
	sigs.k8s.io/yaml v1.2.0
)

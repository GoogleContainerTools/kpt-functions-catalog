module github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gatekeeper

go 1.15

require (
	github.com/open-policy-agent/frameworks/constraint v0.0.0-20210121003109-e55b2bb4cf1c
	github.com/open-policy-agent/gatekeeper v0.0.0-20210409021048-9b5e4cfe5d7e // This is v3.4.0. It has a semver major version of 2 or higher and is not a Go module yet.
	k8s.io/apimachinery v0.19.2
	sigs.k8s.io/kustomize/kyaml v0.10.20
	sigs.k8s.io/yaml v1.2.0
)

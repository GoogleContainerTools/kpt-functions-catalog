module github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gatekeeper

go 1.16

require (
	github.com/open-policy-agent/frameworks/constraint v0.0.0-20210522003146-5c034948ac29
	github.com/open-policy-agent/gatekeeper v0.0.0-20210527161344-e229247f04d1 // This is v3.5.1. It has a semver major version of 2 or higher and is not a Go module yet.
	github.com/spf13/cobra v1.1.1
	k8s.io/apimachinery v0.20.2
	sigs.k8s.io/kustomize/kyaml v0.10.21
	sigs.k8s.io/yaml v1.2.0
)

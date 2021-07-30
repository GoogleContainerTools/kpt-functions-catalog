module github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/starlark

go 1.16

require (
	github.com/qri-io/starlib v0.5.0
	go.starlark.net v0.0.0-20210406145628-7a1108eaa012
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	// We use a unreleased version to pickup https://github.com/kubernetes-sigs/kustomize/pull/4023.
	// We should switch to a released version when the next kyaml is out.
	sigs.k8s.io/kustomize/kyaml v0.11.1-0.20210630191550-02d14d724aa6
	sigs.k8s.io/yaml v1.2.0
)

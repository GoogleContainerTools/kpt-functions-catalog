module github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/policy-controller-validate

go 1.15

require (
	github.com/open-policy-agent/frameworks/constraint v0.0.0-20210317225149-4f80ac172ddf
	github.com/open-policy-agent/gatekeeper v3.0.4-beta.2+incompatible
	k8s.io/apimachinery v0.17.2
	sigs.k8s.io/kustomize/kyaml v0.10.13
	sigs.k8s.io/yaml v1.2.0
)

module github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-namespace

go 1.15

require (
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v1.0.0
	sigs.k8s.io/kustomize/api v0.6.4
	sigs.k8s.io/kustomize/kyaml v0.9.4
	sigs.k8s.io/yaml v1.2.0
)

// TODO: pin to api release
replace sigs.k8s.io/kustomize/api v0.6.4 => sigs.k8s.io/kustomize/api v0.0.0-20201120230733-052a6b4e967b

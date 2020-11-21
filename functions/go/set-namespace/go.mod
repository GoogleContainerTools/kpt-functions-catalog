module main

go 1.14

require (
	github.com/spf13/cobra v1.0.0
	sigs.k8s.io/kustomize/api v0.6.4
	sigs.k8s.io/kustomize/kyaml v0.9.4
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/api v0.6.4 => sigs.k8s.io/kustomize/api v0.0.0-20201120230733-052a6b4e967b

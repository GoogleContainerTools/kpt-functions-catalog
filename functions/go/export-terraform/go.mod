module github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/export-terraform

go 1.16

replace sigs.k8s.io/kustomize/kyaml v0.11.1 => github.com/mengqiy/kustomize/kyaml v0.15.0

require (
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.22.3
	k8s.io/apimachinery v0.22.3
	sigs.k8s.io/kustomize/kyaml v0.11.1
)

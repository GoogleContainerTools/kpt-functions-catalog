module github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/starlark

go 1.16

require (
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	sigs.k8s.io/kustomize/kyaml v0.10.21
	sigs.k8s.io/yaml v1.2.0
)

//replace sigs.k8s.io/kustomize/kyaml v0.10.21 => ../../../../../../sigs.k8s.io/kustomize/kyaml
replace sigs.k8s.io/kustomize/kyaml v0.10.21 => github.com/mengqiy/kustomize/kyaml v0.0.0-20210712171606-e996e89862f1

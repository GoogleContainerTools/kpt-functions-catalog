module github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/export-terraform

go 1.16

replace github.com/GoogleContainerTools/kpt-functions-catalog/thirdparty/kyaml/fnsdk v0.0.0-20220106201853-53d6f3d583fe => ../../../thirdparty/kyaml/fnsdk

require (
	github.com/GoogleContainerTools/kpt-functions-catalog/thirdparty/kyaml/fnsdk v0.0.0-20220106201853-53d6f3d583fe
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.23.1
	k8s.io/apimachinery v0.23.1
)

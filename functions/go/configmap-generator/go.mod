module github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/configmap-generator

go 1.17

require (
	github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/native-config-adaptor v0.0.0-00010101000000-000000000000
	github.com/GoogleContainerTools/kpt-functions-sdk/go/fn v0.0.0-20220723081830-33aee7fe8a2e
	k8s.io/api v0.24.3
)

require (
	github.com/GoogleContainerTools/kpt-functions-sdk/go/api v0.0.0-20220723081830-33aee7fe8a2e // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.8.0 // indirect
	github.com/xlab/treeprint v1.1.0 // indirect
	golang.org/x/net v0.0.0-20220728211354-c7608f3a8462 // indirect
	golang.org/x/sys v0.0.0-20220731174439-a90be440212d // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	honnef.co/go/augeas v0.0.0-20161110001225-ca62e35ed6b8 // indirect
	k8s.io/apimachinery v0.24.3 // indirect
	k8s.io/klog/v2 v2.70.1 // indirect
	k8s.io/kube-openapi v0.0.0-20220627174259-011e075b9cb8 // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9 // indirect
	sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2 // indirect
	sigs.k8s.io/kustomize/kyaml v0.13.8 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
)

replace github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/native-config-adaptor => ../native-config-adaptor
replace github.com/GoogleContainerTools/kpt-functions-sdk/go/fn => ../../../../kpt-functions-sdk/go/fn
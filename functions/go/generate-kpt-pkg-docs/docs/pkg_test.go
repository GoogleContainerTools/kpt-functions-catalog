package docs

import (
	"testing"

	kptfilev1 "github.com/GoogleContainerTools/kpt-functions-sdk/go/pkg/api/kptfile/v1"
	kptutil "github.com/GoogleContainerTools/kpt-functions-sdk/go/pkg/api/util"
	"github.com/stretchr/testify/require"
)

func TestGetFnCfgPaths(t *testing.T) {
	tests := []struct {
		name string
		kf   string
		want []string
	}{
		{
			name: "simple",
			kf: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: kcc-namespace
  annotations:
    blueprints.cloud.google.com/title: Project Namespace Package
    internal.config.kubernetes.io/path: Kptfile
info:
  description: |
    Kubernetes namespace configured for use with Config Connector to manage GCP
    resources in a specific project.
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configPath: setters.yaml
  validators:
    - image: gcr.io/kpt-fn/starlark:v0.3
      configPath: validation.yaml`,
			want: []string{"setters.yaml", "validation.yaml"},
		},
		{
			name: "no pipeline",
			kf: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: kcc-namespace
  annotations:
    blueprints.cloud.google.com/title: Project Namespace Package
    internal.config.kubernetes.io/path: Kptfile
info:
  description: |
    Kubernetes namespace configured for use with Config Connector to manage GCP
    resources in a specific project.
`,
			want: []string{},
		},
		{
			name: "with cm",
			kf: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: kcc-namespace
  annotations:
    blueprints.cloud.google.com/title: Project Namespace Package
    internal.config.kubernetes.io/path: Kptfile
info:
  description: |
    Kubernetes namespace configured for use with Config Connector to manage GCP
    resources in a specific project.
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/apply-setters:v0.1
      configMap:
        foo: bar
  validators:
    - image: gcr.io/kpt-fn/starlark:v0.3
      configPath: validation.yaml`,
			want: []string{"validation.yaml"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			kf := getKfFromStr(t, tt.kf)
			got := getFnCfgPaths(kf)
			require.ElementsMatch(tt.want, got)
		})
	}
}

func getKfFromStr(t *testing.T, k string) *kptfilev1.KptFile {
	t.Helper()
	require := require.New(t)
	kf, err := kptutil.DecodeKptfile(k)
	require.NoError(err)
	return kf
}

package fixpkg

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/copyutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func TestFixV1alpha1ToV1(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	err = copyutil.CopyDir("../../../../testdata/fix/v1alpha1Tov1/nginx-v1alpha1", dir)
	assert.NoError(t, err)
	inout := &kio.LocalPackageReadWriter{
		PackagePath:    dir,
		MatchFilesGlob: append(kio.DefaultMatch, "Kptfile"),
	}
	f := &Fix{}
	err = kio.Pipeline{
		Inputs:  []kio.Reader{inout},
		Filters: []kio.Filter{f},
		Outputs: []kio.Writer{inout},
	}.Execute()
	assert.NoError(t, err)
	diff, err := copyutil.Diff(dir, "../../../../testdata/fix/v1alpha1Tov1/nginx-v1")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(diff.List()))
}

func TestFixV1alpha2ToV1(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	err = copyutil.CopyDir("../../../../testdata/fix/v1alpha2Tov1/nginx-v1alpha2", dir)
	assert.NoError(t, err)
	inout := &kio.LocalPackageReadWriter{
		PackagePath:    dir,
		MatchFilesGlob: append(kio.DefaultMatch, "Kptfile"),
	}
	f := &Fix{}
	err = kio.Pipeline{
		Inputs:  []kio.Reader{inout},
		Filters: []kio.Filter{f},
		Outputs: []kio.Writer{inout},
	}.Execute()
	assert.NoError(t, err)
	diff, err := copyutil.Diff(dir, "../../../../testdata/fix/v1alpha2Tov1/nginx-v1")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(diff.List()))
}

package fixpkg

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/copyutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func TestFix(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	err =  copyutil.CopyDir("../../../../testdata/fix/nginx-v1alpha1", dir)
	inout := &kio.LocalPackageReadWriter{
		PackagePath:     dir,
		MatchFilesGlob:  append(kio.DefaultMatch, "Kptfile"),
	}
	f := &Fix{}
	err = kio.Pipeline{
		Inputs:  []kio.Reader{inout},
		Filters: []kio.Filter{f},
		Outputs: []kio.Writer{inout},
	}.Execute()
	diff, err := copyutil.Diff(dir, "../../../../testdata/fix/nginx-v1alpha2")
	assert.NoError(t, err)
	fmt.Printf("%v", diff.List())
	assert.Equal(t,0, len(diff.List()))
}

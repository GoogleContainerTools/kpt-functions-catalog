package transformer

import (
	"encoding/json"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"sigs.k8s.io/yaml"
	"sort"
	"strings"
)

// implement runner
type SetLabels struct {
	Labels              map[string]string `json:"labels,omitempty"`
	AdditionalFieldSpec []FieldSpec       `json:"additionalLabelFields,omitempty"`
}

// Run the method that every runner need to implement
// TODO: change e2e test cases to match runner: empty config error.
func (r *SetLabels) Run(ctx *fn.Context, functionConfig *fn.KubeObject, items []*fn.KubeObject) {
	if len(r.Labels) == 0 {
		errStr := "failed to configure function: input label list cannot be empty"
		if ctx != nil {
			ctx.Result(errStr, fn.Error)
		}
		fn.Logf(errStr)
		return
	}
	if err := r.Transform(items, ctx); err != nil {
		if ctx != nil {
			ctx.ResultErr(err.Error(), nil)
		}
		fn.Logf(err.Error())
		return
	}
}

func (r *SetLabels) Transform(objects fn.KubeObjects, ctx *fn.Context) error {
	for _, o := range objects {
		if err := visitByDefaultSetting(o, ctx, r.Labels); err != nil {
			return err
		}
		if err := visitBySpecArr(o, r.AdditionalFieldSpec, ctx, r.Labels); err != nil {
			return err
		}
	}
	return nil
}

func visitByDefaultSetting(obj *fn.KubeObject, ctx *fn.Context, newLabels map[string]string) error {
	commonFields := readDefaultFieldSpec()
	return visitBySpecArr(obj, commonFields, ctx, newLabels)
}

func visitBySpecArr(obj *fn.KubeObject, specs []FieldSpec, ctx *fn.Context, newLabels map[string]string) error {
	for _, sp := range specs {
		if (sp.Kind == "" && sp.Version == "" && sp.Group == "") || obj.IsGVK(sp.Group, sp.Version, sp.Kind) {
			// generate msg
			err := updateLabels(obj, newLabels, sp)
			if err != nil {
				return err
			}
			// success, record result
			res, _ := json.Marshal(newLabels)
			msg := "set labels: " + string(res)
			if ctx != nil {
				ctx.ResultInfo(msg, obj)
			}
		}
	}
	return nil
}

// the similar struct esxits in resid.GVK, but there is no function to create an GVK struct without using kyaml
type FieldSpec struct {
	Group              string `json:"group,omitempty" yaml:"group,omitempty"`
	Version            string `json:"version,omitempty" yaml:"version,omitempty"`
	Kind               string `json:"kind,omitempty" yaml:"kind,omitempty"`
	Path               string `json:"path,omitempty" yaml:"path,omitempty"` // seperated by /
	CreateIfNotPresent bool   `json:"create,omitempty" yaml:"create,omitempty"`
}

func updateLabels(o *fn.KubeObject, newLabels map[string]string, spec FieldSpec) error {
	//TODO: should support user configurable field for labels
	basePath := strings.Split(spec.Path, "/")
	keys := make([]string, 0)
	for k := range newLabels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		val := newLabels[key]
		newPath := append(basePath, key)
		_, exist, err := o.NestedString(newPath...)
		if err != nil {
			return err
		}
		if exist || spec.CreateIfNotPresent {
			if err = o.SetNestedString(val, newPath...); err != nil {
				return err
			}
		}
	}
	return nil

}

func readDefaultFieldSpec() []FieldSpec {
	var defaultFieldSpecs []FieldSpec
	yaml.Unmarshal([]byte(commonLabelFieldSpecs), &defaultFieldSpecs)
	return defaultFieldSpecs
}

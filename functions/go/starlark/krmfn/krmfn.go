package krmfn

import (
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const ModuleName = "krmfn.star"

var Module = &starlarkstruct.Module{
	Name: "krmfn",
	Members: starlark.StringDict{
		"match_gvkn": starlark.NewBuiltin("match_gvkn", matchGVKN),
	},
}

func matchGVKN(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var resourceList starlark.Value
	var apiVersion, kind, name string
	if err := starlark.UnpackPositionalArgs("match_gvkn", args, kwargs, 4,
		&resourceList, &apiVersion, &kind, &name); err != nil {
		return nil, err
	}
	obj, err := fn.ParseKubeObject([]byte(resourceList.String()))
	if err != nil {
		return nil, err
	}
	match := obj.GetAPIVersion() == apiVersion && obj.GetKind() == kind && obj.GetName() == name
	return starlark.Bool(match), nil
}

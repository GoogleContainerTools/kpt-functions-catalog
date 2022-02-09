package custom

import (
	"fmt"
	"regexp"

	"sigs.k8s.io/kustomize/api/filters/fieldspec"
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ kio.Filter = Filter{}

type Filter struct {
	ProjectID string
	FsSlice   []types.FieldSpec
}

func (f Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	_, err := kio.FilterAll(yaml.FilterFunc(
		func(node *yaml.RNode) (*yaml.RNode, error) {
			var fns []yaml.Filter
			for _, fs := range f.FsSlice {

				fn := fieldspec.Filter{
					SetValue:  f.updateProjectIDFn(fs.RegexPattern),
					FieldSpec: fs,
				}
				fns = append(fns, fn)
			}
			return node.Pipe(fns...)
		})).Filter(nodes)
	return nodes, err
}

func (f Filter) updateProjectIDFn(regexPath string) filtersutil.SetFn {
	return func(node *yaml.RNode) (err error) {
		if regexPath == "" {
			return node.PipeE(updater{ProjectID: f.ProjectID})
		}
		defer func() {
			// recover from regex panic.
			if recover() != nil {
				err = fmt.Errorf("invalid regex pattern %v", regexPath)
			}
		}()
		re := regexp.MustCompile(regexPath)
		match := re.FindStringSubmatch(node.YNode().Value)
		namedGroup := make(map[string]string)
		for i, name := range re.SubexpNames() {
			if i != 0 && name != "" {
				namedGroup[name] = match[i]
			}
		}
		newProjectID := namedGroup["prefix"] + f.ProjectID + namedGroup["suffix"]
		return node.PipeE(updater{ProjectID: newProjectID})
	}
}

type updater struct {
	ProjectID string
}

func (u updater) Filter(rn *yaml.RNode) (*yaml.RNode, error) {
	return rn.Pipe(yaml.FieldSetter{StringValue: u.ProjectID})
}

package putsettervalues

import (
	"fmt"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/fix/v1"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ kio.Filter = &PutSetterValues{}

// PutSetterValues put the input setter values in Kptfile or functionConfig for setters
type PutSetterValues struct {
	// Setters holds the user provided values for all the setters
	Setters []Setter

	// Results are the results of putting setter values
	Results []*Result

	// filePath file path of resource
	filePath string
}

type Setter struct {
	// Name is the name of the setter
	Name string

	// Value is the input value for setter
	Value string
}

// Result holds result of put setter values operation
type Result struct {
	// FilePath is the file path of the matching field
	FilePath string

	// Name is name of the setter
	Name string

	// Value of the matching field
	Value string
}

// Filter implements Set as a yaml.Filter
func (as *PutSetterValues) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	foundKptfile := false
	for i, node := range nodes {
		if node.GetAnnotations()[kioutil.PathAnnotation] != v1.KptFileName {
			continue
		}

		foundKptfile = true

		kf, err := v1.ReadFile(node)
		if err != nil {
			return nodes, err
		}

		if kf.Pipeline == nil {
			return nodes, nil
		}

		for _, fn := range kf.Pipeline.Mutators {
			if !strings.Contains(fn.Image, "apply-setters") {
				continue
			}
			if fn.ConfigMap != nil {
				for _, setter := range as.Setters {
					val, err := standardizeArrayValue(setter.Value)
					if err != nil {
						return nodes, err
					}
					fn.ConfigMap[setter.Name] = val
					as.Results = append(as.Results, &Result{
						FilePath: v1.KptFileName,
						Name:     setter.Name,
						Value:    setter.Value,
					})
				}
				// convert updated kf to yaml node
				b, err := yaml.Marshal(kf)
				if err != nil {
					return nodes, err
				}
				nodes[i], err = yaml.Parse(string(b))
				if err != nil {
					return nodes, err
				}
			} else if fn.ConfigPath != "" {
				settersConfig, err := settersNode(nodes, fn.ConfigPath)
				if err != nil {
					return nodes, err
				}
				data := settersConfig.GetDataMap()
				for _, setter := range as.Setters {
					val, err := standardizeArrayValue(setter.Value)
					if err != nil {
						return nodes, err
					}
					data[setter.Name] = val
					as.Results = append(as.Results, &Result{
						FilePath: fn.ConfigPath,
						Name:     setter.Name,
						Value:    setter.Value,
					})
				}
				settersConfig.SetDataMap(data)
			}
		}
	}
	if !foundKptfile {
		return nodes, fmt.Errorf(`unable to find "Kptfile" in the package, please ensure "Kptfile" is present in the root directory and specify --include-meta-resources flag`)
	}
	return nodes, nil
}

func settersNode(nodes []*yaml.RNode, path string) (*yaml.RNode, error) {
	for _, node := range nodes {
		np := node.GetAnnotations()[kioutil.PathAnnotation]
		if np == path {
			return node, nil
		}
	}
	return nil, fmt.Errorf(`file %q doesn't exist, please ensure the file specified in "configPath" exists and retry`, path)
}

// Decode decodes the input yaml node into Setter struct
func Decode(rn *yaml.RNode, fcd *PutSetterValues) {
	for k, v := range rn.GetDataMap() {
		fcd.Setters = append(fcd.Setters, Setter{Name: k, Value: v})
	}
}

// standardizeArrayValue returns the folded style array node string
// e.g. for input value [foo, bar], it returns
// - foo
// - bar
func standardizeArrayValue(val string) (string, error) {
	if !strings.HasPrefix(val, "[") {
		// the value is either standardized, or is not a sequence node value
		return val, nil
	}
	vNode, err := yaml.Parse(val)
	if err != nil {
		return val, fmt.Errorf("failed to parse the array node value %q with error %q", val, err.Error())
	}
	if vNode.YNode().Kind == yaml.SequenceNode {
		// standardize array values to folded style
		vNode.YNode().Style = yaml.FoldedStyle
		val, err = vNode.String()
		if err != nil {
			return val, fmt.Errorf("failed to serialize the array node value %q with error %q", val, err.Error())
		}
	}
	return val, nil
}

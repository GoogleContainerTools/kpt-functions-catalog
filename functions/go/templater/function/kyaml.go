package function

import (
	"fmt"
	"log"

	"k8s.io/apimachinery/pkg/api/resource"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	kfilters "sigs.k8s.io/kustomize/kyaml/kio/filters"
)

func KFilter(cfg string) kio.Filter {
	k := kfilters.KFilter{}

	err := k.UnmarshalYAML(func(x interface{}) error {
		// GREP - special case - neet to add cmp funciton and convert to address
		// not sure why only grep isn't address
		// see https://github.com/kubernetes-sigs/kustomize/blob/master/kyaml/kio/filters/filters.go#L21
		grepFilter, ok := k.Filter.(kfilters.GrepFilter)
		if ok {
			grepFilter.Compare = func(a, b string) (int, error) {
				qa, err := resource.ParseQuantity(a)
				if err != nil {
					return 0, fmt.Errorf("%s: %v", a, err)
				}
				qb, err := resource.ParseQuantity(b)
				if err != nil {
					return 0, err
				}

				return qa.Cmp(qb), err
			}
			err := yaml.Unmarshal([]byte(cfg), &grepFilter)
			if err != nil {
				log.Printf("can't unmarshal KFilter grepFilter cfg yaml %s: %v", cfg, err)
				return err
			}
			k.Filter = grepFilter
			return nil
		}

		err := yaml.Unmarshal([]byte(cfg), x)
		if err != nil {
			log.Printf("can't unmarshal KFilter cfg yaml %s: %v", cfg, err)
			return err
		}
		return nil
	})
	if err != nil {
		log.Printf("can't unmarshalYAML KFilter cfg %s: %v", cfg, err)
		return nil
	}

	//log.Printf("return %T %v", k.Filter, k. Filter)
	return k.Filter
}

func KPipe(input []*yaml.RNode, kfilters []interface{}) []*yaml.RNode {
	kfs, err := getKFilters(kfilters)
	if err != nil {
		log.Printf("KPipe: %v", err)
	}

	pb := kio.PackageBuffer{}
	p := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.PackageBuffer{Nodes: input}},
		Filters: kfs,
		Outputs: []kio.Writer{&pb},
	}

	err = p.Execute()
	if err != nil {
		log.Printf("pipeline exec returned error: %v", err)
		return nil
	}
	//log.Printf("KPipe returns %v", pb.Nodes)
	return pb.Nodes
}

func YFilter(cfg string) yaml.Filter {
	y := yaml.YFilter{}

	err := y.UnmarshalYAML(func(x interface{}) error {
		err := yaml.Unmarshal([]byte(cfg), x)
		if err != nil {
			log.Printf("can't unmarshal YFilter cfg yaml %s: %v", cfg, err)
			return err
		}
		return nil
	})

	if err != nil {
		log.Printf("can't unmarshalYAML YFilter cfg %s: %v", cfg, err)
		return nil
	}

	return y.Filter
}

func YPipe(input *yaml.RNode, yfilters []interface{}) *yaml.RNode {
	yfs, err := getYFilters(yfilters)
	if err != nil {
		log.Printf("YPipe: %v", err)
	}
	res, err := input.Pipe(yfs...)
	if err != nil {
		log.Printf("pipe returned error: %v", err)
		return nil
	}
	return res
}

func YValue(input *yaml.RNode) interface{} {
	s, err := input.String()
	if err != nil {
		log.Printf("can't get string for %v: %v", input, err)
		return nil
	}

	var x interface{}
	err = yaml.Unmarshal([]byte(s), &x)
	if err != nil {
		log.Printf("can't unmarshal yaml %s: %v", s, err)
		return nil
	}
	return x
}

type KYFilter struct {
	yfilters []yaml.Filter
}

func NewKYFilter(yfilters []interface{}) kio.Filter {
	yfs, err := getYFilters(yfilters)
	if err != nil {
		log.Printf("KYFilter: %v", err)
	}
	return KYFilter{yfilters: yfs}
}

func (k KYFilter) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, i := range input {
		err := i.PipeE(k.yfilters...)
		if err != nil {
			log.Printf("pipe returned error: %v", err)
			return nil, err
		}
	}
	return input, nil
}

func getYFilters(yfilters []interface{}) ([]yaml.Filter, error) {
	yfs := []yaml.Filter{}
	for i, y := range yfilters {
		yf, ok := y.(yaml.Filter)
		if !ok {
			return nil, fmt.Errorf("has got element %d with unexpected type %T", i, y)
		}
		yfs = append(yfs, yf)
	}
	return yfs, nil
}

func getKFilters(kfilters []interface{}) ([]kio.Filter, error) {
	kfs := []kio.Filter{}
	for i, k := range kfilters {
		kf, ok := k.(kio.Filter)
		if !ok {
			return nil, fmt.Errorf("has got element %d with unexpected type %T", i, k)
		}
		kfs = append(kfs, kf)
	}
	return kfs, nil
}

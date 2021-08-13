// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/inflate-helm-chart/generated"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/inflate-helm-chart/third_party/sigs.k8s.io/kustomize/api/builtins"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

//nolint
func main() {
	asp := HelmChartProcessor{}
	cmd := command.Build(&asp, command.StandaloneEnabled, false)

	cmd.Short = generated.InflateHelmChartShort
	cmd.Long = generated.InflateHelmChartLong
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type HelmChartProcessor struct{}

func (slp *HelmChartProcessor) Process(resourceList *framework.ResourceList) error {
	err := run(resourceList)
	if err != nil {
		resourceList.Result = &framework.Result{
			Name: "inflate-helm-chart",
			Items: []framework.ResultItem{
				{
					Message:  err.Error(),
					Severity: framework.Error,
				},
			},
		}
		return resourceList.Result
	}
	return nil
}

type helmChartInflatorFunction struct {
	kyaml.ResourceMeta `json:",inline" yaml:",inline"`
	plugins []builtins.HelmChartInflationGeneratorPlugin
}

func (f *helmChartInflatorFunction) Config(rn *kyaml.RNode) error {
	y, err := rn.String()
	if err != nil {
		return fmt.Errorf("cannot get YAML from RNode: %w", err)
	}
	kind, err := f.getKind(rn)
	if err != nil {
		return err
	}
	switch kind {
	case "InflateHelmChart":
		err = f.ConfigHelmArgs(nil, []byte(y))
		if err != nil {
			return err
		}
	case "ConfigMap":
		dataMap := rn.GetDataMap()
		bytes, err := kyaml.Marshal(dataMap)
		if err != nil {
			return err
		}
		err = f.ConfigMapArgs(bytes)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("`functionConfig` must be `ConfigMap` or `InflateHelmChart`")
	}
	return nil
}

func (f *helmChartInflatorFunction) Run(items []*kyaml.RNode) ([]*kyaml.RNode, error) {
	resmapFactory := builtins.NewResMapFactory()
	resMap, err := resmapFactory.NewResMapFromRNodeSlice(items)
	if err != nil {
		return nil, err
	}
	var rm resmap.ResMap
	for _, p := range f.plugins {
		rm, err = p.Generate()
		if err != nil {
			return nil, fmt.Errorf("failed to run generator: %w", err)
		}

		// check for duplicates for idempotency
		for i := range items {
			resources := rm.Resources()
			for r := range resources {
				it := &resource.Resource{RNode: *items[i]}
				if resources[r].CurId() == it.CurId() {
					// don't attempt to add a resource with the same ID
					err := rm.Remove(resources[r].CurId())
					if err != nil {
						return items, err
					}
				}
			}
		}

		err = resMap.AppendAll(rm)
		if err != nil {
			return nil, fmt.Errorf("failed to add generated resource: %w", err)
		}
	}
	return resMap.ToRNodeSlice(), nil
}

func run(resourceList *framework.ResourceList) error {
	var fn helmChartInflatorFunction
	err := fn.Config(resourceList.FunctionConfig)
	if err != nil {
		return fmt.Errorf("failed to configure function: %w", err)
	}
	resourceList.Items, err = fn.Run(resourceList.Items)
	if err != nil {
		return fmt.Errorf("failed to run function: %w", err)
	}
	return nil
}

func (f *helmChartInflatorFunction) getKind(rn *kyaml.RNode) (string, error) {
	meta, err := rn.GetMeta()
	if err != nil {
		return "", err
	}
	return meta.Kind, nil
}

func (f *helmChartInflatorFunction) ConfigHelmArgs(
	_ *resmap.PluginHelpers, c []byte) (err error) {
	args := &builtins.HelmArgs{}
	if err = kyaml.Unmarshal(c, args); err != nil {
		return
	}
	for _, helmChart := range args.HelmCharts {
		p := builtins.HelmChartInflationGeneratorPlugin{
			HelmGlobals: args.HelmGlobals,
			HelmChart:   helmChart,
		}
		if err := p.ValidateArgs(); err != nil {
			return err
		}
		f.plugins = append(f.plugins, p)
	}

	return nil
}

func (f *helmChartInflatorFunction) ConfigMapArgs(
	bytes []byte) (err error) {
	var p builtins.HelmChartInflationGeneratorPlugin
	err = kyaml.Unmarshal(bytes, &p)
	if err != nil {
		return err
	}

	if err := p.ValidateArgs(); err != nil {
		return err
	}
	f.plugins = append(f.plugins, p)
	return nil
}

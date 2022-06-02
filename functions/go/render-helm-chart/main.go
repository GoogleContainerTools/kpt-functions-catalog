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

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/render-helm-chart/generated"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/render-helm-chart/third_party/sigs.k8s.io/kustomize/api/builtins"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	fnConfigKind = "RenderHelmChart"
	configMap    = "ConfigMap"
)

//nolint
func main() {
	asp := HelmChartProcessor{}
	cmd := command.Build(&asp, command.StandaloneEnabled, false)

	cmd.Short = generated.RenderHelmChartShort
	cmd.Long = generated.RenderHelmChartLong
	cmd.Example = generated.RenderHelmChartExamples
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
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

type helmChartInflatorFunction struct {
	kyaml.ResourceMeta `json:",inline" yaml:",inline"`
	plugins            []builtins.HelmChartInflationGeneratorPlugin
}

func (f *helmChartInflatorFunction) Config(rn *kyaml.RNode) error {
	var err error
	switch rn.GetKind() {
	case fnConfigKind:
		err = f.RenderHelmChartArgs(rn)
	case configMap:
		err = f.ConfigMapArgs(rn.GetDataMap())
	default:
		err = fmt.Errorf("`functionConfig` must be `%s` or `%s`", configMap, fnConfigKind)
	}
	return err
}

func (f *helmChartInflatorFunction) Run(items []*kyaml.RNode) ([]*kyaml.RNode, error) {
	resmapFactory := builtins.NewResMapFactory()
	resMap, err := resmapFactory.NewResMapFromRNodeSlice(items)
	if err != nil {
		return nil, err
	}
	var rm resmap.ResMap
	for _, p := range f.plugins {
		err := p.ConfigureAuth(items)
		if err != nil {
			return nil, err
		}
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

func (f *helmChartInflatorFunction) RenderHelmChartArgs(rn *kyaml.RNode) (err error) {
	y, err := rn.String()
	if err != nil {
		return fmt.Errorf("cannot get YAML from RNode: %w", err)
	}
	c := []byte(y)
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
	m map[string]string) (err error) {
	var p builtins.HelmChartInflationGeneratorPlugin
	if val, ok := m["chartHome"]; ok {
		p.ChartHome = val
	}
	if val, ok := m["configHome"]; ok {
		p.ConfigHome = val
	}
	if val, ok := m["name"]; ok {
		p.ChartArgs.Name = val
	}
	if val, ok := m["version"]; ok {
		p.ChartArgs.Version = val
	}
	if val, ok := m["repo"]; ok {
		p.ChartArgs.Repo = val
	}
	if val, ok := m["releaseName"]; ok {
		p.TemplateOptions.ReleaseName = val
	}
	if val, ok := m["namespace"]; ok {
		p.TemplateOptions.Namespace = val
	}
	if val, ok := m["nameTemplate"]; ok {
		p.TemplateOptions.NameTemplate = val
	}
	if val, ok := m["includeCRDs"]; ok {
		if val == "true" {
			p.TemplateOptions.IncludeCRDs = true
		}
	}
	if val, ok := m["skipTests"]; ok {
		if val == "true" {
			p.TemplateOptions.SkipTests = true
		}
	}
	if val, ok := m["valuesFile"]; ok {
		p.TemplateOptions.ValuesFiles = []string{val}
	}
	if err := p.ValidateArgs(); err != nil {
		return err
	}
	f.plugins = append(f.plugins, p)
	return nil
}

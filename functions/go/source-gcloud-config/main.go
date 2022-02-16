// Copyright 2022 Google LLC
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

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/source-gcloud-generator/generated"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/source-gcloud-generator/generator"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

type Processor struct{}

func (p *Processor) Process(resourceList *framework.ResourceList) error {
	err := func() error {
		gen := generator.NewGcloudConfigGenerator()
		updated, err := gen.Generate(resourceList.Items)
		if err != nil {
			return err
		}
		resourceList.Items = updated
		return nil
	}()
	if err != nil {
		resourceList.Results = framework.Results{
			&framework.Result{
				Message:  err.Error(),
				Severity: framework.Error,
			},
		}
		return resourceList.Results
	}
	return nil
}

func main() {
	cmd := command.Build(&Processor{}, command.StandaloneEnabled, false)
	cmd.Short = generated.SourceGcloudConfigShort
	cmd.Long = generated.SourceGcloudConfigLong
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

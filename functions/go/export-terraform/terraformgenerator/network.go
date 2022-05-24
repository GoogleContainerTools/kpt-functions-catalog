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

package terraformgenerator

import (
	sdk "github.com/GoogleContainerTools/kpt-functions-catalog/thirdparty/kyaml/fnsdk"
)

type firewallAllow struct {
	Ports    []string `yaml:"ports,omitempty"`
	Protocol string   `yaml:"protocol"`
}

func (resource *terraformResource) GetFirewallAllowPortsProtocol() []firewallAllow {
	var firewallAllows []firewallAllow
	found, err := resource.Item.Get(&firewallAllows, "spec", "allow")
	if !found || err != nil {
		sdk.Logf("unable to find allowed FW ports/protocol in %s (found = %t): %s\n", resource.Name, found, err)
	}
	return firewallAllows
}

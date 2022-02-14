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
package exec

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"gopkg.in/ini.v1"
)

var domainRegex = regexp.MustCompile(`\S+@(\S+)`)

// For testing
var GetGcloudContextFn = GetGcloudContext

// GetGcloudContext executes `gcloud` commands to get default gcloud config value.
func GetGcloudContext() (map[string]string, error) {
	var projectID, zone, region, account, domain, orgID string
	gcloudCfg, err := ListGcloudConfig()
	if err != nil || gcloudCfg == "" {
		return nil, err
	}

	// Read the INI formatted gcloud config.
	// DO NOT USE github.com/go-gcfg/gcfg which has a nicer function UI but does not support INI with underscore.
	// gcloud config `disable_usage_reporting` breaks.
	iniCfg, err := ini.LoadSources(ini.LoadOptions{Loose: true}, []byte(gcloudCfg))
	if err != nil {
		return nil, err
	}
	core := iniCfg.Section("core")
	if core != nil {
		projectID = core.Key("project").String()
		account = core.Key("account").String()
	}
	compute := iniCfg.Section("compute")
	if compute != nil {
		zone = compute.Key("zone").String()
		region = compute.Key("region").String()
	}

	// Query cloud server to get the OrganizationID which the project ID belongs to.
	if projectID != "" {
		orgID, err = getGcloudOrgID(projectID)
		if err != nil {
			return nil, err
		}
	}
	if account != "" {
		// e.g. account `NAME@COMPANY.com` has matching domain `COMPANY.com`
		matches := domainRegex.FindStringSubmatch(account)
		if len(matches) < 2 {
			// Skip if cannot parse `domain` from gcloud `core/account`
		} else {
			domain = matches[1]
		}
	}
	return map[string]string{
		"namespace": projectID,
		"projectID": projectID,
		"zone":      zone,
		"region":    region,
		"domain":    domain,
		"orgID":     orgID,
	}, nil

}

// ListGcloudConfig runs `gcloud config list` to read local default gcloud configuration.
func ListGcloudConfig() (string, error) {
	var cmdOut bytes.Buffer
	cmd := exec.Command("gcloud", "config", "list")
	cmd.Stdout = &cmdOut
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("unable to run `gcloud config list` %v", err.Error())
	}
	return cmdOut.String(), nil
}

// getGcloudOrgID queries cloud server to get the `Organization ID`.
func getGcloudOrgID(projectID string) (string, error) {
	var buf, err, out bytes.Buffer
	cmdListAncestors := exec.Command("gcloud", "projects", "get-ancestors",
		projectID, "--format=get(id)")
	cmdListAncestors.Stdout = &buf
	cmdListAncestors.Stderr = &err
	if e := cmdListAncestors.Run(); e != nil {
		return "", e
	}
	if err.Len() > 0 {
		return "", fmt.Errorf(err.String())
	}
	cmdOrgID := exec.Command("tail", "-1")
	cmdOrgID.Stdin = &buf
	cmdListAncestors.Stderr = &err
	cmdOrgID.Stdout = &out
	if e := cmdOrgID.Run(); e != nil {
		return "", e
	}
	if err.Len() > 0 {
		return "", fmt.Errorf(err.String())
	}
	raw := out.String()
	return strings.TrimSpace(raw), nil
}

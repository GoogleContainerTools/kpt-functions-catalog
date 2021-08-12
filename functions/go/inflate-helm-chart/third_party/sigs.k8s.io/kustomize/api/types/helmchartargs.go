// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

type HelmGlobals struct {
	// ChartHome is a file path to a directory containing a subdirectory for
	// each chart to be included in the output. The default value of this field 
	// is "tmp/charts".
	// At runtime, the function will look for the chart under {ChartHome}. If it
	// is there, the function will use it as found. If it is not there, the
	// function will attempt to pull it and put it in {ChartHome}.
	// When run as a container function, local directories must be mounted into
	// the container in order for the function to use them.
	// If the function needs to pull the helm chart while running in a container,
	// ChartHome MUST start with "tmp/".
	ChartHome string `json:"chartHome,omitempty" yaml:"chartHome,omitempty"`

	// ConfigHome defines a value that the function should pass to helm via
	// the HELM_CONFIG_HOME environment variable. The function doesn't attempt
	// to read or write this directory.
	// If omitted, {tmpDir}/helm is used, where {tmpDir} is some temporary
	// directory created by the function for the benefit of helm.
	// Likewise, the function sets
	//   HELM_CACHE_HOME={ConfigHome}/.cache
	//   HELM_DATA_HOME={ConfigHome}/.data
	// for the helm subprocess.
	ConfigHome string `json:"configHome,omitempty" yaml:"configHome,omitempty"`
}

type HelmChart struct {
	// Name is the name of the chart, e.g. 'minecraft'.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Version is the version of the chart, e.g. '3.1.3'
	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	// Repo is a URL locating the chart on the internet.
	// This is the argument to helm's  `--repo` flag, e.g.
	// `https://itzg.github.io/minecraft-server-charts`.
	Repo string `json:"repo,omitempty" yaml:"repo,omitempty"`

	// ReleaseName replaces RELEASE-NAME in chart template output,
	// making a particular inflation of a chart unique with respect to
	// other inflations of the same chart in a cluster. It's the first
	// argument to the helm `install` and `template` commands, i.e.
	//   helm install {RELEASE-NAME} {chartName}
	//   helm template {RELEASE-NAME} {chartName}
	// If omitted, the flag --generate-name is passed to 'helm template'.
	ReleaseName string `json:"releaseName,omitempty" yaml:"releaseName,omitempty"`

	// Namespace set the target namespace for a release. It is .Release.Namespace
	// in the helm template
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	// ValuesFile is local file path to a values file to use _instead of_
	// the default values that accompanied the chart.
	// The default values are in '{ChartHome}/{Name}/values.yaml'.
	ValuesFile string `json:"valuesFile,omitempty" yaml:"valuesFile,omitempty"`

	// ValuesInline holds value mappings specified directly,
	// rather than in a separate file.
	ValuesInline map[string]interface{} `json:"valuesInline,omitempty" yaml:"valuesInline,omitempty"`

	// ValuesMerge specifies how to treat ValuesInline with respect to Values.
	// Legal values: 'merge', 'override', 'replace'.
	// Defaults to 'override'.
	ValuesMerge string `json:"valuesMerge,omitempty" yaml:"valuesMerge,omitempty"`

	// IncludeCRDs specifies if Helm should also generate CustomResourceDefinitions.
	// Defaults to 'false'.
	IncludeCRDs bool `json:"includeCRDs,omitempty" yaml:"includeCRDs,omitempty"`
}

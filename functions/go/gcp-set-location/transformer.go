package main

import (
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gcp-set-location/fieldspec"
	"sigs.k8s.io/kustomize/api/filters/fsslice"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	ZoneKey   = "zone"
	RegionKey = "region"
)

type Transformer struct {
	region          string
	zone            string
	regionFieldSpec types.FsSlice
	zoneFieldSpec   types.FsSlice
}
type NoOp struct {
}

func (e *NoOp) Error() string {
	return fmt.Sprintf("neither `region` nor `zone` is given")
}

func (t *Transformer) Config(fnConfigNode *yaml.RNode) error {
	data := fnConfigNode.GetDataMap()
	if data == nil {
		return fmt.Errorf("missing `data` field in `ConfigMap` FunctionConfig")
	}
	region, ok := data[RegionKey]
	if ok && region != "" {
		if deactive, ok := CloudRegions[region]; !ok || !deactive {
			return fmt.Errorf("unknown region %v", region)
		}
	}
	t.region = region
	zone, ok := data[ZoneKey]
	if ok && zone != "" {
		if deactive, ok := CloudZones[zone]; !ok || !deactive {
			return fmt.Errorf("unknown zone %v", zone)
		}
	}
	t.zone = zone

	if t.zone == "" && t.region == "" {
		return &NoOp{}
	}

	// Enumerate project field specs
	var rfs CustomFieldSpec
	if err := yaml.Unmarshal([]byte(fieldspec.RegionFieldSpecs), &rfs); err != nil {
		return err
	}
	t.regionFieldSpec = rfs.RegionFieldSpecs
	var zfs CustomFieldSpec
	if err := yaml.Unmarshal([]byte(fieldspec.ZoneFieldSpecs), &zfs); err != nil {
		return err
	}
	t.zoneFieldSpec = zfs.ZoneFieldSpecs

	return nil
}

func (p *Transformer) Transform(m resmap.ResMap) error {
	return m.ApplyFilter(CustomFieldSpecFilter{
		Region:        p.region,
		regionFsSlice: p.regionFieldSpec,
		Zone:          p.zone,
		zoneFsSlice:   p.zoneFieldSpec,
	})
}

var _ kio.Filter = CustomFieldSpecFilter{}

func (f CustomFieldSpecFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	_, err := kio.FilterAll(yaml.FilterFunc(f.filter)).Filter(nodes)
	return nodes, err
}

type CustomFieldSpec struct {
	RegionFieldSpecs []types.FieldSpec `json:"region,omitempty" yaml:"region,omitempty"`
	ZoneFieldSpecs   []types.FieldSpec `json:"zone,omitempty" yaml:"zone,omitempty"`
}

type CustomFieldSpecFilter struct {
	Region        string
	Zone          string
	regionFsSlice types.FsSlice `json:"region,omitempty" yaml:"region,omitempty"`
	zoneFsSlice   types.FsSlice `json:"zone,omitempty" yaml:"zone,omitempty"`
}

func (f CustomFieldSpecFilter) filter(node *yaml.RNode) (*yaml.RNode, error) {
	if f.Zone != "" {
		f := fsslice.Filter{
			FsSlice:  f.zoneFsSlice,
			SetValue: f.updateZoneFn,
		}
		if err := node.PipeE(f); err != nil {
			return nil, err
		}
	}
	if f.Region != "" {
		f := fsslice.Filter{
			FsSlice:  f.regionFsSlice,
			SetValue: f.updateRegionFn,
		}
		if err := node.PipeE(f); err != nil {
			return nil, err
		}
	}
	return node, nil
}

func (f CustomFieldSpecFilter) updateRegionFn(node *yaml.RNode) error {
	return node.PipeE(updater{location: f.Region})
}

func (f CustomFieldSpecFilter) updateZoneFn(node *yaml.RNode) error {
	return node.PipeE(updater{location: f.Zone})
}

type updater struct {
	location string
}

func (u updater) Filter(rn *yaml.RNode) (*yaml.RNode, error) {
	return rn.Pipe(yaml.FieldSetter{StringValue: u.location})
}

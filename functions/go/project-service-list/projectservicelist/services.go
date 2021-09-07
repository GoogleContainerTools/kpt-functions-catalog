package projectservicelist

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Service represents a valid GCP Service
type Service struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		ResourceID string `yaml:"resourceID"`
		ProjectRef struct {
			External string `yaml:"external,omitempty"`
		} `yaml:"projectRef,omitempty"`
	} `yaml:"spec"`
}

// getServicesList creates a slice of services from service strings
func getServicesList(name string, inputServices []string, projectID string) ([]Service, error) {
	var services []Service
	for _, serviceName := range inputServices {
		if len(strings.Split(serviceName, ".")) < 3 {
			return nil, fmt.Errorf("invalid service specified: %s", serviceName)
		}
		serviceShortName := strings.Split(serviceName, ".")[0]
		service := Service{APIVersion: serviceUsageAPIVersion, Kind: serviceUsageKind}
		service.Metadata.Name = fmt.Sprintf("%s-%s", name, serviceShortName)
		service.Spec.ResourceID = serviceName
		if projectID != "" {
			service.Spec.ProjectRef.External = projectID
		}
		services = append(services, service)
	}
	return services, nil
}

// createService generates a Service RNode from Service struct
func createService(s Service) (*yaml.RNode, error) {
	yml, err := yaml.Marshal(s)
	if err != nil {
		return nil, err
	}
	r, err := yaml.Parse(string(yml))
	if err != nil {
		return nil, err
	}
	return r, nil
}

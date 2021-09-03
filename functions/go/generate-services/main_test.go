package main

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const annotatedProjectKrm string = `apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
kind: Project
metadata:
  name: host-project-id # {"$kpt-set":"host-project-id"}
  namespace: projects # {"$kpt-set":"projects-namespace"}
  annotations:
    cnrm.cloud.google.com/auto-create-network: "false"
    cnrm.cloud.google.com/organization-id: "123456789012" # {"$kpt-set":"org-id"}
spec: {}`

const annotatedProjectService string = `apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: projects-cloudresourcemanager
  namespace: gcp-services
  annotations:
    cnrm.cloud.google.com/project-id: projects
spec:
  resourceID: cloudresourcemanager.googleapis.com`

const annotatedProjectServiceDiffProject string = `apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
kind: Service
metadata:
  name: projects-cloudresourcemanager
  namespace: gcp-services
  annotations:
    cnrm.cloud.google.com/project-id: projects-01
spec:
  resourceID: cloudresourcemanager.googleapis.com`

const annotatedPolicyKrm string = `apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMPolicyMember
metadata:
  name: service-project-id-container-network-user
  namespace: projects
  annotations:
    cnrm.cloud.google.com/project-id: host-project-id
spec: {}`

func TestShouldRun(t *testing.T) {
	serviceMap, err := loadServiceMap(defaultServiceMapPath)
	if err != nil {
		t.Fatal("failed to load ServiceMap:\n", err)
	}
	item, err := yaml.Parse(annotatedProjectKrm)
	if err != nil {
		t.Fatal("not a valid KRM resource\n", err)
	}
	itemMeta, err := item.GetMeta()
	if err != nil {
		t.Fatal("GetMeta Errored:\n", err)
	}
	svcList, supported, err := getRequiredServices(serviceMap, itemMeta)
	if err != nil {
		t.Fatal("getRequiredServices Errored:\n", err)
	}
	if !supported {
		t.Fatal("kind.group SHOULD be supported.\n", annotatedProjectKrm)
	}
	expectedSvcList := []string{"cloudresourcemanager.googleapis.com"}
	if !reflect.DeepEqual(svcList, expectedSvcList) {
		t.Fatalf("Unexpected result. Got %v. Wanted %v.\n", svcList, expectedSvcList)
	}

	item, err = yaml.Parse(annotatedPolicyKrm)
	if err != nil {
		t.Fatal("not a valid KRM resource\n", err)
	}
	itemMeta, err = item.GetMeta()
	if err != nil {
		t.Fatal("GetMeta Errored:\n", err)
	}
	svcList, supported, err = getRequiredServices(serviceMap, itemMeta)
	if err != nil {
		t.Fatal("getRequiredServices Errored:\n", err)
	}
	if supported {
		t.Fatal("kind.group SHOULD NOT be supported.\n", annotatedPolicyKrm)
	}
	expectedSvcList = nil
	if !reflect.DeepEqual(svcList, expectedSvcList) {
		t.Fatalf("Unexpected result. Got %v. Wanted %v.\n", svcList, expectedSvcList)
	}
}

func TestExistingServicePresent(t *testing.T) {
	serviceMap, err := loadServiceMap(defaultServiceMapPath)
	if err != nil {
		t.Fatal("failed to load ServiceMap:\n", err)
	}

	itemProject, err := yaml.Parse(annotatedProjectKrm)
	if err != nil {
		t.Fatal("not a valid KRM resource\n", err)
	}

	itemService, err := yaml.Parse(annotatedProjectService)
	if err != nil {
		t.Fatal("not a valid KRM resource\n", err)
	}

	resourceList := MakeResourceListFromNodes(itemProject, itemService)

	existingServices, err := createExistingServicesMap(resourceList)
	if err != nil {
		t.Fatal("unable to create existing services map\n", err)
	}

	projectServices, err := createProjectServicesMap(resourceList, serviceMap, existingServices)
	if err != nil {
		t.Fatal("unable to create project services map\n", err)
	}

	// require 0 services because there is already a Service resource for "projects"
	require.Equal(t, 0, len(projectServices["projects"]))
}

func TestExistingServiceNotPresent(t *testing.T) {
	serviceMap, err := loadServiceMap(defaultServiceMapPath)
	if err != nil {
		t.Fatal("failed to load ServiceMap:\n", err)
	}

	itemProject, err := yaml.Parse(annotatedProjectKrm)
	if err != nil {
		t.Fatal("not a valid KRM resource\n", err)
	}

	resourceList := MakeResourceListFromNodes(itemProject)

	existingServices, err := createExistingServicesMap(resourceList)
	if err != nil {
		t.Fatal("unable to create existing services map\n", err)
	}

	projectServices, err := createProjectServicesMap(resourceList, serviceMap, existingServices)
	if err != nil {
		t.Fatal("unable to create project services map\n", err)
	}

	// require 1 service because there is no Service resource for "projects"
	require.Equal(t, 1, len(projectServices["projects"]))
}

func TestExistingServicePresentInDifferentProject(t *testing.T) {
	serviceMap, err := loadServiceMap(defaultServiceMapPath)
	if err != nil {
		t.Fatal("failed to load ServiceMap:\n", err)
	}

	itemProject, err := yaml.Parse(annotatedProjectKrm)
	if err != nil {
		t.Fatal("not a valid KRM resource\n", err)
	}

	itemServiceDiffProject, err := yaml.Parse(annotatedProjectServiceDiffProject)
	if err != nil {
		t.Fatal("not a valid KRM resource\n", err)
	}

	resourceList := MakeResourceListFromNodes(itemProject, itemServiceDiffProject)

	existingServices, err := createExistingServicesMap(resourceList)
	if err != nil {
		t.Fatal("unable to create existing services map\n", err)
	}

	projectServices, err := createProjectServicesMap(resourceList, serviceMap, existingServices)
	if err != nil {
		t.Fatal("unable to create project services map\n", err)
	}

	// require 1 service because there is one Service resource but it is for
	// a different project i.e. "other-projects"
	require.Equal(t, 1, len(projectServices["projects"]))
}

func MakeResourceListFromNodes(resources ...*yaml.RNode) *framework.ResourceList {
	var resourceList framework.ResourceList
	resourceList.Items = resources
	return &resourceList
}

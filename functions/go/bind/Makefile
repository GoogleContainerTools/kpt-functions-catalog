# GCP project to use for development
export GCP_PROJECT_ID ?= \
	$(if $(shell which gcloud),$(shell gcloud config get-value project), \
	gcr.io/kpt-fn)
export IMAGE_REPO ?= gcr.io/$(GCP_PROJECT_ID)
export IMAGE_TAG ?= latest

build:
	KO_DOCKER_REPO=ko.local ko build -B --tags=${IMAGE_TAG} .

push:
	KO_DOCKER_REPO=${IMAGE_REPO} ko build -B --tags=${IMAGE_TAG} .


# GCP project to use for development
export GCP_PROJECT_ID ?= $(shell gcloud config get-value project)
export IMAGE_REPO ?= gcr.io/$(GCP_PROJECT_ID)
export IMAGE_TAG ?= latest

build:
	KO_DOCKER_REPO=ko.local ko build -B --tags=${IMAGE_TAG} .

push:
	KO_DOCKER_REPO=${IMAGE_REPO} ko build -B --tags=${IMAGE_TAG} .


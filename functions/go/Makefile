# Copyright 2020 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
SHELL=/bin/bash
GOPATH := $(shell go env GOPATH)
TAG := latest

# Edit this list to contain all go functions
FUNCTIONS := \
	set-namespace

# Targets for running all function tests
FUNCTION_TESTS := $(patsubst %,%-TEST,$(FUNCTIONS))
# Targets to build functions
FUNCTION_BUILDS := $(patsubst %,%-BUILD,$(FUNCTIONS))
# Targets to build functions with dev tag
FUNCTION_DEV_BUILDS := $(patsubst %,%-BUILD-DEV,$(FUNCTIONS))
# Targets to push images
FUNCTION_PUSH := $(patsubst %,%-PUSH,$(FUNCTIONS))
# Current function name used by individual function targets
CURRENT_FUNCTION := unknown

.DEFAULT_GOAL := help
.PHONY: help
help: ## Print this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: $(FUNCTIONS)
$(FUNCTIONS):
	$(MAKE) CURRENT_FUNCTION=$@ func-verify

.PHONY: test
test: $(FUNCTION_TESTS) ## Run unit tests for all functions

.PHONY: $(FUNCTION_TESTS)
$(FUNCTION_TESTS):
	$(MAKE) CURRENT_FUNCTION=$(subst -TEST,,$@) func-test

.PHONY: build
build: $(FUNCTION_BUILDS) ## Build all function images. 'TAG' env is used to specify tag. 'latest' will be used if not set.

.PHONY: $(FUNCTION_BUILDS)
$(FUNCTION_BUILDS):
	$(MAKE) CURRENT_FUNCTION=$(subst -BUILD,,$@) func-build

.PHONY: build-dev
build-dev: $(FUNCTION_DEV_BUILDS) ## Build all function images with tag 'dev'. This is used for local tests.

.PHONY: $(FUNCTION_DEV_BUILDS)
$(FUNCTION_DEV_BUILDS):
	$(MAKE) CURRENT_FUNCTION=$(subst -BUILD-DEV,,$@) TAG=dev func-build

.PHONY: push
push: $(FUNCTION_PUSH) ## Push images to registry. WARN: This operation should only be done in CI environment.

.PHONY: $(FUNCTION_PUSH)
$(FUNCTION_PUSH):
	$(MAKE) CURRENT_FUNCTION=$(subst -PUSH,,$@) func-push

# Recipes for individual function
.PHONY: func-fix func-vet func-fmt func-test func-lint \
	func-build func-push func-verify
func-verify: func-fix func-vet func-fmt func-test func-lint

func-fix:
	cd $(CURRENT_FUNCTION) && go fix ./...

func-fmt:
	cd $(CURRENT_FUNCTION) && go fmt ./...

func-lint:
	(which $(GOPATH)/bin/golangci-lint || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.32.2)
	cd $(CURRENT_FUNCTION) && $(GOPATH)/bin/golangci-lint run ./...

func-test:
	cd $(CURRENT_FUNCTION) && go test -cover ./...

func-vet:
	cd $(CURRENT_FUNCTION) && go vet ./...

func-build: func-verify
	@echo Building image for $(CURRENT_FUNCTION)...
	@( \
		image_tag=$${TAG:-latest}; \
		prefix="release-go-functions-"; \
		[[ "$${image_tag}" = "$${prefix}"* ]] && image_tag="$${image_tag#$$prefix}"; \
		image_name=gcr.io/kpt-functions/"$${CURRENT_FUNCTION}"; \
		image="$${image_name}":"$${image_tag}"; \
		docker build -t "$${image}" -t "$${image_name}" -f "$${CURRENT_FUNCTION}"/Dockerfile "$${CURRENT_FUNCTION}"; \
	)

func-push:
	@echo Pushing image $(CURRENT_FUNCTION)...
	@( \
		image_tag=$${TAG:-latest}; \
		prefix="release-go-functions-"; \
		[[ "$${image_tag}" = "$${prefix}"* ]] && image_tag="$${image_tag#$$prefix}"; \
		image_name=gcr.io/kpt-functions/"$${CURRENT_FUNCTION}"; \
		image="$${image_name}":"$${image_tag}"; \
		docker push "$${image_name}:latest"; \
		docker push "$${image}"; \
	)

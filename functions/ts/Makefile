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
TAG ?= dev
CURATED_GCR = gcr.io/kpt-fn
CONTRIB_GCR = gcr.io/kpt-fn-contrib

.DEFAULT_GOAL := help
.PHONY: help
help: ## Print this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test build npm-ci push

# Edit this list to contain all of the curated ts functions
CURATED_FUNCTIONS := \
	kubeval

# Edit this list to contain all of the non-curated ts functions
CONTRIB_FUNCTIONS := \
	helm-inflator \
	istioctl-analyze \
	sops

FUNCTIONS := $(CURATED_FUNCTIONS) $(CONTRIB_FUNCTIONS)

# Targets for running all function tests
FUNCTION_TESTS := $(patsubst %,%-TEST,$(FUNCTIONS))

FUNCTION_CHECKLICENSES := $(patsubst %,%-CHECKLICENSES,$(FUNCTIONS))
# Targets to build curated functions
CURATED_FUNCTION_BUILDS := $(patsubst %,%-CURATED-BUILD,$(CURATED_FUNCTIONS))
# Targets to build non-curated functions
CONTRIB_FUNCTION_BUILDS := $(patsubst %,%-CONTRIB-BUILD,$(CONTRIB_FUNCTIONS))
# Targets to push curated function images
CURATED_FUNCTION_PUSH := $(patsubst %,%-CURATED-PUSH,$(CURATED_FUNCTIONS))
# Targets to push non-curated function images
CONTRIB_FUNCTION_PUSH := $(patsubst %,%-CONTRIB-PUSH,$(CONTRIB_FUNCTIONS))
# Targets to run npm ci
FUNCTION_NPM_CI := $(patsubst %,%-NPM-CI,$(FUNCTIONS))
# Current function name used by individual function targets
CURRENT_FUNCTION ?= unknown

npm-ci: $(FUNCTION_NPM_CI) ## Run npm ci command

.PHONY: $(FUNCTION_NPM_CI)
$(FUNCTION_NPM_CI):
	$(MAKE) CURRENT_FUNCTION=$(subst -NPM-CI,,$@) func-npm-ci

test: $(FUNCTION_TESTS) ## Run unit tests for all functions

.PHONY: $(FUNCTION_TESTS)
$(FUNCTION_TESTS):
	$(MAKE) CURRENT_FUNCTION=$(subst -TEST,,$@) func-test

build: $(CURATED_FUNCTION_BUILDS) $(CONTRIB_FUNCTION_BUILDS) ## Build all function images. Variable 'TAG' is used to specify tag. 'dev' will be used if not set.

.PHONY: $(CURATED_FUNCTION_BUILDS)
$(CURATED_FUNCTION_BUILDS):
	$(MAKE) CURRENT_FUNCTION=$(subst -CURATED-BUILD,,$@) TAG=$(TAG) DEFAULT_GCR=$(CURATED_GCR) func-build

.PHONY: $(CONTRIB_FUNCTION_BUILDS)
$(CONTRIB_FUNCTION_BUILDS):
	$(MAKE) CURRENT_FUNCTION=$(subst -CONTRIB-BUILD,,$@) TAG=$(TAG) DEFAULT_GCR=$(CONTRIB_GCR) func-build

push: $(CURATED_FUNCTION_PUSH) $(CONTRIB_FUNCTION_PUSH) ## Push images to registry. WARN: This operation should only be done in CI environment.

.PHONY: $(CURATED_FUNCTION_PUSH)
$(CURATED_FUNCTION_PUSH):
	$(MAKE) CURRENT_FUNCTION=$(subst -CURATED-PUSH,,$@) TAG=$(TAG) DEFAULT_GCR=$(CURATED_GCR) func-push

.PHONY: $(CONTRIB_FUNCTION_PUSH)
$(CONTRIB_FUNCTION_PUSH):
	$(MAKE) CURRENT_FUNCTION=$(subst -CONTRIB-PUSH,,$@) TAG=$(TAG) DEFAULT_GCR=$(CONTRIB_GCR) func-push

check-licenses: $(FUNCTION_CHECKLICENSES) ## Run license checker for source files

.PHONY: $(FUNCTION_CHECKLICENSES)
$(FUNCTION_CHECKLICENSES):
	$(MAKE) CURRENT_FUNCTION=$(subst -CHECKLICENSES,,$@) func-check-licenses

# Recipes for individual function
.PHONY: func-test func-check-licenses func-build func-push func-npm-ci
func-npm-ci:
	cd $(CURRENT_FUNCTION) && npm ci

func-test: func-npm-ci
	cd $(CURRENT_FUNCTION) && npm test

func-check-licenses: func-npm-ci
	cd $(CURRENT_FUNCTION) && npm run check-licenses

func-build: func-test
	../../scripts/ts-function-release.sh build

func-push:
	../../scripts/ts-function-release.sh push

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
TAG := latest

.DEFAULT_GOAL := help
.PHONY: help
help: ## Print this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test build build-dev push

test: ## Run unit tests for all functions
	cd go && $(MAKE) test
	cd ts && $(MAKE) test

build: ## Build all function images. Variable 'TAG' is used to specify tag. 'latest' will be used if not set.
	cd go && $(MAKE) TAG=$(TAG) build
	cd ts && $(MAKE) TAG=$(TAG) build

build-dev: ## Build all function images with tag 'dev'. This is used for local tests.
	cd go && $(MAKE) build-dev
	cd ts && $(MAKE) build-dev

push: ## Push images to registry. WARN: This operation should only be done in CI environment.
	cd go && $(MAKE) push
	cd ts && $(MAKE) push

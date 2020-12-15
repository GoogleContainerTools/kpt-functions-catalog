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
TAG := dev

.DEFAULT_GOAL := help
.PHONY: help
help: ## Print this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test build npm-ci push
npm-ci:
	npm ci

test: npm-ci ## Run unit tests for all functions
	npm test

build: test ## Build all function images. Variable 'TAG' is used to specify tag. 'dev' will be used if not set.
	npm run kpt:docker-build -- --tag $(TAG)

lint-license: ## Run license linter for source files
	npm run lint-license

push: ## Push images to registry. WARN: This operation should only be done in CI environment.
	npm run kpt:docker-push -- --tag=$(TAG)

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
TAG := unstable

.DEFAULT_GOAL := help
.PHONY: help
help: ## Print this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test unit-test e2e-test build push

unit-test: unit-test-go unit-test-ts ## Run unit tests for all functions

unit-test-go: ## Run unit tests for Go functions
	cd functions/go && $(MAKE) test
	cd contrib/functions/go && $(MAKE) test

unit-test-ts: ## Run unit tests for TS functions
	cd functions/ts && $(MAKE) test
	cd contrib/functions/ts && $(MAKE) test

e2e-test: ## Run all e2e tests
	cd tests && $(MAKE) TAG=$(TAG) test

test: unit-test e2e-test ## Run all unit tests and e2e tests

check-licenses:
	cd functions/ts && $(MAKE) check-licenses
	cd functions/go && $(MAKE) check-licenses
	cd contrib/functions/go && $(MAKE) check-licenses
	cd contrib/functions/ts && $(MAKE) check-licenses

verify-docs:
	go install github.com/monopole/mdrip@v1.0.2
	(cd scripts/patch_reader/ && go build -o patch_reader .)
	scripts/verify-docs.py

build: ## Build all function images. Variable 'TAG' is used to specify tag. 'dev' will be used if not set.
	cd functions/go && $(MAKE) TAG=$(TAG) build
	cd functions/ts && $(MAKE) TAG=$(TAG) build
	cd contrib/functions/go && $(MAKE) TAG=$(TAG) build
	cd contrib/functions/ts && $(MAKE) TAG=$(TAG) build

push: ## Push images to registry. WARN: This operation should only be done in CI environment.
	cd functions/go && $(MAKE) push
	cd functions/ts && $(MAKE) push
	cd contrib/functions/go && $(MAKE) push
	cd contrib/functions/ts && $(MAKE) push

site-generate: ## Collect function branches and generate a catalog of their examples and documentation using kpt next.
	rm -rf ./site/*/ && mkdir site/contrib
	(cd scripts/generate_catalog/ && go run . ../.. ../../site)

site-run: ## Run the site locally.
	make site-generate
	./scripts/run-site.sh

site-check: ## Test site for broken catalog links.
	make site-run
	./scripts/check-site.sh

update-function-docs: ## Update documentation for a function release branch
	(cd scripts/update_function_docs/ && go build -o update_function_docs .)
	RELEASE_BRANCH=$(RELEASE_BRANCH) ./scripts/update_function_docs/update_function_docs

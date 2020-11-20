GOPATH := $(shell go env GOPATH)

FUNCTIONS := \
	set-namespace

CURRENT_FUNCTION := unknown

.PHONY: fix vet fmt test lint all
.PHONY: $(FUNCTIONS)

all: $(FUNCTIONS)

$(FUNCTIONS):
	$(MAKE) CURRENT_FUNCTION=$@ function

function: fix vet fmt test lint

fix:
	cd $(CURRENT_FUNCTION) && go fix ./...

fmt:
	cd $(CURRENT_FUNCTION) && go fmt ./...

lint:
	(which $(GOPATH)/bin/golangci-lint || go get github.com/golangci/golangci-lint/cmd/golangci-lint)
	cd $(CURRENT_FUNCTION) && $(GOPATH)/bin/golangci-lint run ./...

test:
	cd $(CURRENT_FUNCTION) && go test -cover ./...

vet:
	cd $(CURRENT_FUNCTION) && go vet ./...

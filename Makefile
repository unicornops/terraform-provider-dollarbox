GO ?= go
GOFMT ?= gofmt

.DEFAULT_GOAL := test

build:
	$(GO) build -v ./...

install: build
	$(GO) install -v .

fmt:
	$(GOFMT) -s -w .

lint:
	golangci-lint run

test:
	$(GO) test -v -race -cover ./...

testacc:
	TF_ACC=1 $(GO) test -v -cover -timeout 120m ./...

tidy:
	$(GO) mod tidy

vet:
	$(GO) vet ./...

.PHONY: build install fmt lint test testacc tidy vet

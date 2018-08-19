BINARY = email_forwarder
TARGET = ./cmd/$(BINARY)
LINT_PACKAGES = $(shell go list -f '{{.Dir}}' ./...)
GOFLAGS ?=
COVERPROFILES = *.coverprofile
COVERPROFILE = goverprofile
CWD = $(shell pwd)

PACKAGES = github.com/modocache/gover\
	github.com/mattn/goveralls\
	github.com/golang/lint/golint\
	github.com/golang/dep/cmd/dep

# General
all: build

build:
	go build $(GOFLAGS) $(TARGET)

lint:
	go vet ./...
	golint $(LINT_PACKAGES)
	gofmt -d -s $(LINT_PACKAGES)

setup:
	go get -u $(PACKAGES)

clean:
	rm -f $(COVERPROFILE) $(COVERPROFILES) $(BINARY)

.PHONY: all build lint setup clean

# Testing
test:
	go list -f '{{if len .TestGoFiles}}"go test -coverprofile='$(CWD)'/{{.Name}}.coverprofile {{.ImportPath}}"{{end}}' ./... | xargs -L 1 sh -c

$(COVERPROFILES): test

$(COVERPROFILE): $(COVERPROFILES)
	gover . $@

cover: $(COVERPROFILE)
	go tool cover -html=$< -o cover.html
	gocov convert $< | gocov report

.PHONY: test cover

# CI
ci_prepare: setup
	dep ensure

coveralls: $(COVERPROFILE)
	goveralls -coverprofile=$< -service=travis-ci

ci: test coveralls lint


.PHONY: ci_prepare coveralls ci

# Extra
lambda:
	GOOS=linux GOFLAGS='-ldflags="-s -w"' make build && zip $(BINARY).zip $(BINARY) ef.json

all: test vet build
.PHONY: all

download: go.sum
	go mod download
.PHONY: download

mod-tidy:
	go mod tidy
.PHONY: mod-tidy

generate: download
	# When not running tests or vetting, we don't want to generate mocks (vektra/mockery) as it's quite slow
	go generate -run='//go:generate go run github.com/([^v]|v[^e]|ve[^k]|vek[^t])' -x ./...
ifeq (, $(findstring test,$(MAKECMDGOALS))$(findstring vet,$(MAKECMDGOALS)))
else
	go generate -x ./...
endif
.PHONY: generate

vet: download generate
	gofmt -d -e -s .
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck ./...
.PHONY: vet

test: download generate
	go test -v ./...
.PHONY: test

TEST_OUT_FILE ?= report.xml
COV_OUT_FILE ?= coverage.out

test-ci: download generate
	go run gotest.tools/gotestsum --junitfile $(TEST_OUT_FILE) -- -coverprofile=$(COV_OUT_FILE) -v ./...
.PHONY: test-ci

coverage-ci:
	go run github.com/mattn/goveralls -coverprofile=$(COV_OUT_FILE) -service=github
.PHONY: coverage-ci

TARGET ?= missing

build: download generate
	go build ./cmd/$(TARGET)
.PHONY: build

## Docker
DOCKER_IMAGE_PREFIX ?= local
DOCKER_TAGS ?= latest
DOCKER_PLATFORM ?= linux/$(shell uname -m)
DOCKER_IMAGE ?= $(DOCKER_IMAGE_PREFIX)/$(TARGET)
DOCKER_IMAGE_TAG ?= $(DOCKER_IMAGE):$(DOCKER_TAGS)

DOCKER_PLATFORM_FLAG =
ifneq ($(DOCKER_PLATFORM),)
	DOCKER_PLATFORM_FLAG = --platform $(DOCKER_PLATFORM)
endif

KO_ARGS ?= --local

build-docker:
	KO_DOCKER_REPO=$(DOCKER_IMAGE) ko build --sbom=none --bare --tags $(DOCKER_TAGS) $(DOCKER_PLATFORM_FLAG) $(KO_ARGS) ./cmd/$(TARGET)
.PHONY: build-docker

build-push-docker:
	make build-docker KO_ARGS="--push"
.PHONY: build-push-docker

ARGS ?=
DOCKER_ARGS ?=

run: download generate
	go run ./cmd/$(TARGET) $(ARGS)
.PHONY: run

run-docker: build-docker
	docker run -it --rm $(DOCKER_ARGS) $(DOCKER_IMAGE_TAG) $(ARGS)
.PHONY: run-docker

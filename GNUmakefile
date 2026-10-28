default: build

BINARY_NAME=terraform-provider-openmetadata
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
# OpenMetadata server version the acceptance tests run against. The CI workflow
# reads .github/openmetadata-versions.txt directly and docker-compose.yml takes
# the same value through OPENMETADATA_VERSION, so bump the file, not this
# variable, when moving to a new release.
OPENMETADATA_VERSION ?= $(shell tr -d '[:space:]' < .github/openmetadata-versions.txt)
# Compose implementation. The stack depends on health- and completion-based
# conditions, which need Docker Compose v2 or podman-compose 1.6+. Override with
# `make testacc-docker COMPOSE=podman-compose` when Docker is unavailable.
COMPOSE ?= docker compose

build:
	go build -o bin/$(BINARY_NAME) -ldflags "-X main.version=$(VERSION)" .

test:
	go test -v -cover ./internal/...

test-race:
	go test -v -race ./internal/...

# Unit coverage only: `TestAcc*` needs TF_ACC and a reachable OpenMetadata
# server, so it is skipped to keep the profile independent of the environment.
coverage:
	go test -count=1 -skip '^TestAcc' -covermode=atomic -coverprofile=coverage.out ./internal/...
	@echo
	@./scripts/coverage-summary.sh coverage.out

coverage-html: coverage
	go tool cover -html=coverage.out

# Unit coverage (coverage.out) plus the acceptance profile written by
# docker-compose.yml (coverage-acceptance.out) as one number.
coverage-merge:
	./scripts/merge-coverage.sh coverage-union.out coverage.out coverage-acceptance.out
	@echo
	@./scripts/coverage-summary.sh coverage-union.out

# Regenerate the coverage badge and floor statement in README.md from
# .github/coverage-baseline.txt; `make coverage-readme-check` fails when the
# committed README has drifted from the baseline.
coverage-readme:
	./scripts/coverage-readme.sh

coverage-readme-check:
	./scripts/coverage-readme.sh --check

# Fail when coverage.out is below the floor in .github/coverage-baseline.txt or
# a package floor in .github/coverage-floors.txt. Acceptance coverage is
# reported separately and is not gated: it depends on Docker being available.
coverage-gate:
	./scripts/coverage-gate.sh coverage.out

lint:
	golangci-lint run --config .golangci.yml ./...

lint-fix:
	golangci-lint run --config .golangci.yml --fix ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

generate:
	go generate ./...

validate-examples:
	./scripts/validate-examples.sh

testacc:
	TF_ACC=1 go test -v -timeout 30m ./internal/...

# Runs the acceptance suite in the compose network, against the pinned server.
testacc-docker:
	OPENMETADATA_VERSION=$(OPENMETADATA_VERSION) $(COMPOSE) run --rm test

testacc-down:
	OPENMETADATA_VERSION=$(OPENMETADATA_VERSION) $(COMPOSE) down -v

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/caberdo/openmetadata/$(VERSION)/$$(go env GOOS)_$$(go env GOARCH)
	cp bin/$(BINARY_NAME) ~/.terraform.d/plugins/registry.terraform.io/caberdo/openmetadata/$(VERSION)/$$(go env GOOS)_$$(go env GOARCH)/$(BINARY_NAME)

release-check:
	goreleaser check --config .github/goreleaser.yml

# Build every published platform into dist/ without publishing or signing.
release-snapshot:
	goreleaser release --snapshot --clean --skip=publish,sign --config .github/goreleaser.yml

.PHONY: default build test test-race coverage coverage-html coverage-merge coverage-readme coverage-readme-check coverage-gate lint lint-fix fmt vet generate validate-examples testacc testacc-docker testacc-down install release-check release-snapshot

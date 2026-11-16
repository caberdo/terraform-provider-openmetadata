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
# The shared tooling for the caberdo providers: the coverage gate, merge and
# README targets, and example validation. Keep this pin equal to the `uses:`
# refs in .github/workflows/*.yml: one version for the local targets and the
# reusable workflows, so they cannot drift apart.
TFKIT ?= go run github.com/caberdo/terraform-provider-kit/cmd/tfkit@v0.13.0

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
	@go tool cover -func=coverage.out

coverage-html: coverage
	go tool cover -html=coverage.out

# Unit coverage (coverage.out) plus the acceptance profile written by
# docker-compose.yml (coverage-acceptance.out) as one number.
coverage-merge:
	$(TFKIT) coverage merge --output coverage-union.out coverage.out coverage-acceptance.out
	@echo
	@go tool cover -func=coverage-union.out

# Regenerate the coverage badge and floor statement in README.md from
# .github/coverage-baseline.txt; `make coverage-readme-check` fails when the
# committed README has drifted from the baseline. tfkit renders the sentence
# around `make coverage-gate` itself, so the local and CI verdicts agree.
coverage-readme:
	$(TFKIT) coverage readme

coverage-readme-check:
	$(TFKIT) coverage readme --check

# Fail when coverage.out is below the floor in .github/coverage-baseline.txt or
# a package floor in .github/coverage-floors.txt. Acceptance coverage is
# reported separately and is not gated: it depends on Docker being available.
coverage-gate:
	$(TFKIT) coverage gate --profile coverage.out

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

# Validate every Terraform example against the provider built from this tree.
# tfkit builds into a temporary directory; set TFKIT_PLUGIN_DIR (or pass
# --plugin-dir) to pin that directory, so parallel runs do not overwrite each
# other's build. Terraform missing skips the check unless REQUIRE_TERRAFORM=1.
validate-examples:
	$(TFKIT) examples validate --provider openmetadata

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

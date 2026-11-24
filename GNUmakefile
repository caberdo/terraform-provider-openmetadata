default: build

BINARY_NAME=terraform-provider-openmetadata
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
# Supported OpenMetadata server versions, ordered ascending: the first entry is
# the supported floor and the last is the newest release. CI reads
# .github/openmetadata-versions.txt directly and docker-compose.yml takes the
# selected value through OPENMETADATA_VERSION, so bump the file, not these
# variables, when moving to a new release. Override the window for a quick local
# check: OPENMETADATA_VERSIONS="2.0.8 2.0.9" make testacc-matrix
OPENMETADATA_VERSIONS ?= $(shell sed -e 's/\#.*//' -e 's/[[:space:]]//g' .github/openmetadata-versions.txt | grep -v '^$$')
# The version a plain `make testacc-docker` runs: the newest in the file.
OPENMETADATA_VERSION ?= $(lastword $(OPENMETADATA_VERSIONS))
# Compose implementation. The stack depends on health- and completion-based
# conditions, which need Docker Compose v2 or podman-compose 1.6+. Override with
# `make testacc-docker COMPOSE=podman-compose` when Docker is unavailable.
COMPOSE ?= docker compose
# The shared tooling for the caberdo providers: the coverage gate, merge and
# README targets, and example validation. Keep this pin equal to the `uses:`
# refs in .github/workflows/*.yml: one version for the local targets and the
# reusable workflows, so they cannot drift apart.
TFKIT ?= go run github.com/caberdo/terraform-provider-kit/cmd/tfkit@v0.17.0

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

# Runs the acceptance suite in the compose network, against the newest version
# in the supported window unless OPENMETADATA_VERSION names another one.
testacc-docker:
	OPENMETADATA_VERSION=$(OPENMETADATA_VERSION) $(COMPOSE) run --rm test

# Runs the acceptance suite against every version in the supported window, the
# way CI does. Each version gets its own compose project so a failed leg cannot
# leave containers, volumes or port bindings behind for the next one; the
# overall exit status is non-zero when any version failed.
testacc-matrix:
	@test -n "$(OPENMETADATA_VERSIONS)" || { \
		echo "OPENMETADATA_VERSIONS is empty; check .github/openmetadata-versions.txt" >&2; \
		exit 1; \
	}
	@results=$$(mktemp); status=0; \
	for v in $(OPENMETADATA_VERSIONS); do \
		project=$$(echo "$$v" | tr '.' '-'); \
		echo "==> Acceptance tests against OpenMetadata $$v"; \
		if COMPOSE_PROJECT_NAME=openmetadata-acc-$$project \
			$(MAKE) --no-print-directory testacc-docker OPENMETADATA_VERSION=$$v; then \
			echo "PASS OpenMetadata $$v" >> "$$results"; \
		else \
			echo "FAIL OpenMetadata $$v" >> "$$results"; status=1; \
		fi; \
		COMPOSE_PROJECT_NAME=openmetadata-acc-$$project OPENMETADATA_VERSION=$$v \
			$(COMPOSE) down -v >/dev/null 2>&1 || true; \
	done; \
	echo; echo "==> Acceptance matrix summary"; cat "$$results"; rm -f "$$results"; \
	exit $$status

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

.PHONY: default build test test-race coverage coverage-html coverage-merge coverage-readme coverage-readme-check coverage-gate lint lint-fix fmt vet generate validate-examples testacc testacc-docker testacc-matrix testacc-down install release-check release-snapshot

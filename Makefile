GREEN  := $(shell printf '\033[32m')
RESET  := $(shell printf '\033[0m')

GORELEASER ?= goreleaser

OTELCOL_BUILDER_VERSION ?= 0.144.0
OTELCOL_BUILDER_DIR ?= ${HOME}/bin
OTELCOL_BUILDER ?= ${OTELCOL_BUILDER_DIR}/ocb

DISTRIBUTIONS ?= $(shell echo $(notdir $(wildcard ./distributions/*)) | tr " " ",") # outputs comma separated directories names

## Run the full CI pipeline (check + build)
.PHONY: ci
ci: check build

## Run all checks (test + goreleaser config validation)
.PHONY: check
check: test ensure-goreleaser-up-to-date

## Build the Tulip distribution binary
.PHONY: build
build: ocb
	@./scripts/build.sh -d "${DISTRIBUTIONS}" -b ${OTELCOL_BUILDER}

## Run tests against the built distribution
.PHONY: test
test: build
	@./test/test-all.sh -d "${DISTRIBUTIONS}"

## Generate all artifacts (sources + goreleaser config)
.PHONY: generate
generate: generate-sources generate-goreleaser

## Generate the goreleaser configuration file
.PHONY: generate-goreleaser
generate-goreleaser:
	@./scripts/generate-goreleaser.sh -d "${DISTRIBUTIONS}"

## Generate Go source code from manifest.yaml (no compilation)
.PHONY: generate-sources
generate-sources: ocb
	@./scripts/build.sh -d "${DISTRIBUTIONS}" -s true -b ${OTELCOL_BUILDER}

DOCKER_IMAGE ?= tulip:local
DOCKER_GOARCH ?= $(shell go env GOARCH)

## Build a local Docker image (cross-compiles for linux)
.PHONY: docker
docker: ocb
	@for dist in $$(echo "${DISTRIBUTIONS}" | tr "," " "); do \
		echo "Building $${dist} for linux/${DOCKER_GOARCH}..."; \
		CGO_ENABLED=0 GOOS=linux GOARCH=${DOCKER_GOARCH} \
			./scripts/build.sh -d "$${dist}" -b ${OTELCOL_BUILDER}; \
		cp distributions/$${dist}/_build/$${dist} distributions/$${dist}/$${dist}; \
		docker build -t ${DOCKER_IMAGE} distributions/$${dist}/; \
		rm -f distributions/$${dist}/$${dist}; \
		echo "Image built: ${DOCKER_IMAGE}"; \
	done

## Test the Docker image (build + run + send traces + verify)
.PHONY: docker-test
docker-test: docker
	@for dist in $$(echo "${DISTRIBUTIONS}" | tr "," " "); do \
		DOCKER_IMAGE=${DOCKER_IMAGE} ./test/test-docker.sh -d "$${dist}"; \
	done

## Build a release snapshot with goreleaser (dry-run)
.PHONY: goreleaser-verify
goreleaser-verify: goreleaser
	@${GORELEASER} release --snapshot --clean

## Verify goreleaser config is up to date with templates
.PHONY: ensure-goreleaser-up-to-date
ensure-goreleaser-up-to-date: generate-goreleaser
	@git diff -s --exit-code distributions/*/.goreleaser.yaml || (echo "Check failed: The goreleaser templates have changed but the .goreleaser.yamls haven't. Run 'make generate-goreleaser' and update your PR." && exit 1)

## Install the OpenTelemetry Collector Builder (ocb) if missing
.PHONY: ocb
ocb:
ifeq (, $(shell command -v ocb 2>/dev/null))
	@{ \
	[ ! -x '$(OTELCOL_BUILDER)' ] || exit 0; \
	set -e ;\
	os=$$(uname | tr A-Z a-z) ;\
	machine=$$(uname -m) ;\
	[ "$${machine}" != x86 ] || machine=386 ;\
	[ "$${machine}" != x86_64 ] || machine=amd64 ;\
	echo "Installing ocb ($${os}/$${machine}) at $(OTELCOL_BUILDER_DIR)";\
	mkdir -p $(OTELCOL_BUILDER_DIR) ;\
	curl -sfLo $(OTELCOL_BUILDER) "https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/cmd%2Fbuilder%2Fv$(OTELCOL_BUILDER_VERSION)/ocb_$(OTELCOL_BUILDER_VERSION)_$${os}_$${machine}" ;\
	chmod +x $(OTELCOL_BUILDER) ;\
	}
else
OTELCOL_BUILDER=$(shell command -v ocb)
endif

## Check that goreleaser is installed
.PHONY: goreleaser
goreleaser:
	@{ \
		if ! command -v '$(GORELEASER)' >/dev/null 2>/dev/null; then \
			echo >&2 '$(GORELEASER) command not found. Please install goreleaser. https://goreleaser.com/install/'; \
			exit 1; \
		fi \
	}

REMOTE?=git@github.com:ollygarden/tulip.git
## Create and push a signed release tag (requires TAG=vX.Y)
.PHONY: push-tags
push-tags:
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	@echo "Adding tag ${TAG}"
	@git tag -a ${TAG} -s -m "Version ${TAG}"
	@echo "Pushing tag ${TAG}"
	@git push ${REMOTE} ${TAG}

## Display help for all targets
.PHONY: help
help:
	@awk '/^.PHONY: / { \
		msg = match(lastLine, /^## /); \
			if (msg) { \
				cmd = substr($$0, 9, 100); \
				msg = substr(lastLine, 4, 1000); \
				printf "  ${GREEN}%-30s${RESET} %s\n", cmd, msg; \
			} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

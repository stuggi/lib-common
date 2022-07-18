# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 0.0.1

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.9.0

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.24

.PHONY: all
all: build

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

##@ Build

.PHONY: build
build: fmt vet ## Build a test lib-common binary.
	go build -o lib-common

# CI tools repo for running tests
CI_TOOLS_REPO := https://github.com/openstack-k8s-operators/openstack-k8s-operators-ci
CI_TOOLS_REPO_DIR = $(shell pwd)/CI_TOOLS_REPO
.PHONY: get-ci-tools
get-ci-tools:
	if [ -d  "$(CI_TOOLS_REPO_DIR)" ]; then \
		echo "Ci tools exists"; \
		pushd "$(CI_TOOLS_REPO_DIR)"; \
		git pull --rebase; \
		popd; \
	else \
		git clone $(CI_TOOLS_REPO) "$(CI_TOOLS_REPO_DIR)"; \
	fi

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: generate                                      
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./modules/..."

# Run go fmt against code
gofmt: get-ci-tools
	$(CI_TOOLS_REPO_DIR)/test-runner/gofmt.sh
	$(CI_TOOLS_REPO_DIR)/test-runner/gofmt.sh ./modules/archive
	$(CI_TOOLS_REPO_DIR)/test-runner/gofmt.sh ./modules/common
	$(CI_TOOLS_REPO_DIR)/test-runner/gofmt.sh ./modules/database

# Run go vet against code
govet: get-ci-tools
	$(CI_TOOLS_REPO_DIR)/test-runner/govet.sh
	$(CI_TOOLS_REPO_DIR)/test-runner/govet.sh ./modules/archive
	$(CI_TOOLS_REPO_DIR)/test-runner/govet.sh ./modules/common
	$(CI_TOOLS_REPO_DIR)/test-runner/govet.sh ./modules/database

# Run go test against code
gotest: get-ci-tools
	$(CI_TOOLS_REPO_DIR)/test-runner/gotest.sh
	$(CI_TOOLS_REPO_DIR)/test-runner/gotest.sh ./modules/archive
	$(CI_TOOLS_REPO_DIR)/test-runner/gotest.sh ./modules/common
	$(CI_TOOLS_REPO_DIR)/test-runner/gotest.sh ./modules/database

# Run golangci-lint test against code
golangci: get-ci-tools
	$(CI_TOOLS_REPO_DIR)/test-runner/golangci.sh

# Run go lint against code
golint: get-ci-tools
	PATH=$(GOBIN):$(PATH); $(CI_TOOLS_REPO_DIR)/test-runner/golint.sh
	PATH=$(GOBIN):$(PATH); $(CI_TOOLS_REPO_DIR)/test-runner/golint.sh ./modules/common
	PATH=$(GOBIN):$(PATH); $(CI_TOOLS_REPO_DIR)/test-runner/golint.sh ./modules/database

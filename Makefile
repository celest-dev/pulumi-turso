PROJECT_NAME := Pulumi Turso Provider

PACK             := turso
PACKDIR          := sdk
PROJECT          := github.com/celest-dev/pulumi-turso
NODE_MODULE_NAME := @celest-dev/pulumi-turso
NUGET_PKG_NAME   := Celest.Pulumi.Turso

PROVIDER        := pulumi-resource-${PACK}
PROVIDER_PATH   := provider
VERSION_PATH    := ${PROVIDER_PATH}.Version

SCHEMA_FILE     := provider/cmd/pulumi-resource-turso/schema.json

GOPATH			:= $(shell go env GOPATH)

WORKING_DIR     := $(shell pwd)
EXAMPLES_DIR    := ${WORKING_DIR}/examples/yaml
TESTPARALLELISM := 4

OS    := $(shell uname)
SHELL := /bin/bash

# Override during CI using `make [TARGET] PROVIDER_VERSION=""` or by setting a PROVIDER_VERSION environment variable
# Local & branch builds will just use this fixed default version unless specified
PROVIDER_VERSION ?= 0.0.1-dev
# Use this normalised version everywhere rather than the raw input to ensure consistency.
VERSION_GENERIC = $(shell pulumictl convert-version --language generic --version "$(PROVIDER_VERSION)")

openapi::
	@echo "Generating OpenAPI client"
	@cd $(TMPDIR); \
		curl -sLo openapi.json "https://raw.githubusercontent.com/tursodatabase/turso-docs/refs/heads/main/api-reference/openapi.json"; \
		cp openapi.json $(WORKING_DIR)/provider/internal/tursoclient/openapi.json
	cd $(WORKING_DIR)/provider && go generate ./...

ensure::
	cd provider && go mod tidy
	cd sdk && go mod tidy
	cd tests && go mod tidy

bin/$(PROVIDER)::
	cd provider && go build -o $(WORKING_DIR)/bin/${PROVIDER} -ldflags "-X ${PROJECT}/${VERSION_PATH}=${VERSION_GENERIC}" $(PROJECT)/${PROVIDER_PATH}/cmd/$(PROVIDER)

provider:: bin/$(PROVIDER)

provider_debug::
	cd provider && go build -o $(WORKING_DIR)/bin/${PROVIDER} -gcflags="all=-N -l" -ldflags "-X ${PROJECT}/${VERSION_PATH}=${VERSION_GENERIC}" $(PROJECT)/${PROVIDER_PATH}/cmd/$(PROVIDER)

$(SCHEMA_FILE):: bin/$(PROVIDER)
	pulumi package get-schema $(WORKING_DIR)/bin/${PROVIDER} | jq 'del(.version)' > $(SCHEMA_FILE)

schema:: $(SCHEMA_FILE)

test_provider::
	cd tests && go test -short -v -count=1 -cover -timeout 2h -parallel ${TESTPARALLELISM} ./...

go_sdk:: $(SCHEMA_FILE)
	rm -rf sdk/go
	pulumi package gen-sdk $(SCHEMA_FILE) --language go --version "${VERSION_GENERIC}"
	cd sdk && go mod tidy

nodejs_sdk:: $(SCHEMA_FILE)
	rm -rf sdk/nodejs
	pulumi package gen-sdk $(SCHEMA_FILE) --language nodejs --version "${VERSION_GENERIC}"

build_nodejs_sdk:: nodejs_sdk
	cd ${PACKDIR}/nodejs/ && \
		yarn install && \
		yarn run tsc
	cp README.md LICENSE ${PACKDIR}/nodejs/package.json ${PACKDIR}/nodejs/yarn.lock ${PACKDIR}/nodejs/bin/ 2>/dev/null || true

gen_examples: gen_go_example gen_nodejs_example

gen_%_example:
	rm -rf ${WORKING_DIR}/examples/$*
	pulumi convert \
		--cwd ${WORKING_DIR}/examples/yaml \
		--logtostderr \
		--generate-only \
		--non-interactive \
		--language $* \
		--out ${WORKING_DIR}/examples/$*

define pulumi_login
    export PULUMI_CONFIG_PASSPHRASE=asdfqwerty1234; \
    pulumi login --local;
endef

up::
	$(call pulumi_login) \
	cd ${EXAMPLES_DIR} && \
	pulumi stack init dev && \
	pulumi stack select dev && \
	pulumi config set name dev && \
	pulumi up -y

down::
	$(call pulumi_login) \
	cd ${EXAMPLES_DIR} && \
	pulumi stack select dev && \
	pulumi destroy -y && \
	pulumi stack rm dev -y

.PHONY: build

build:: provider go_sdk nodejs_sdk
all:: build

# Required for the codegen action that runs in pulumi/pulumi
only_build:: build

lint::
	for DIR in "provider" "sdk" "tests" ; do \
		pushd $$DIR && golangci-lint run -c ../.golangci.yml --timeout 10m && popd ; \
	done

install::
	cp $(WORKING_DIR)/bin/${PROVIDER} ${GOPATH}/bin

GO_TEST 	 := go test -v -count=1 -cover -timeout 2h -parallel ${TESTPARALLELISM}

test_all:: test_provider
	cd tests/sdk/go && $(GO_TEST) ./...

install_go_sdk::
	#target intentionally blank

install_nodejs_sdk:: build_nodejs_sdk
	-yarn unlink --cwd $(WORKING_DIR)/sdk/nodejs/bin
	yarn link --cwd $(WORKING_DIR)/sdk/nodejs/bin
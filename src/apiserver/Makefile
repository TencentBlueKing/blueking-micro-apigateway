ifdef VERSION
    VERSION=${VERSION}
else
    VERSION=$(shell git describe --always 2>/dev/null || echo "--")
endif

GITCOMMIT=$(shell git rev-parse HEAD 2>/dev/null || echo "--")
BUILDTIME=${shell date +%Y-%m-%dT%H:%M:%S%z}

LDFLAGS="-X github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/version.Version=${VERSION} \
	-X github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/version.GitCommit=${GITCOMMIT} \
	-X github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/version.BuildTime=${BUILDTIME}"

.PHONY: init
init:
	## 安装 golangci-lint 二进制
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(LOCALBIN) v2.7.2
	## 安装 gofumpt 二进制
	GOBIN=$(LOCALBIN) go install mvdan.cc/gofumpt@latest
	## 安装 goimports-reviser 二进制
	GOBIN=$(LOCALBIN) go install github.com/incu6us/goimports-reviser/v3@latest
	## 安装 swag 二进制
	GOBIN=$(LOCALBIN) go install github.com/swaggo/swag/v2/cmd/swag@latest
	## 安装 golines 二进制
	GOBIN=$(LOCALBIN) go install github.com/segmentio/golines@latest

# go mod tidy
.PHONY: tidy
tidy:
	go mod tidy


.PHONY: dep
dep: tidy
	go mod vendor

# build executable binary
.PHONY: build
build: tidy doc
	CGO_ENABLED=0 go build  -ldflags ${LDFLAGS} -o  apiserver ./main.go

# run unittest
.PHONY: test
test: tidy
	go test ./... -gcflags="all=-N -l" -covermode=count -coverprofile .coverage.cov


integration-test:
	cd tests/integration && docker-compose down && docker-compose up --abort-on-container-exit

# generate swagger api doc with swag
.PHONY: doc
doc:
	GOBIN=$(LOCALBIN) go install github.com/swaggo/swag/v2/cmd/swag@latest
	$(SWAG) fmt
	$(SWAG) init --parseDependency --parseDepth 3

# build docker image
.PHONY: docker-build
docker-build:
	docker build --build-arg VERSION=`git describe --tags --abbrev=0` --build-arg COMMIT=`git rev-parse HEAD` -f ./Dockerfile -t bk-micro-apigw:dev .

.PHONY: fmt
fmt:
	$(GOLINTER) fmt

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint: fmt vet
	$(GOLINTER) run

.PHONY: mock
mock:
	go generate ./...

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
GOLINES ?= $(LOCALBIN)/golines
GOFUMPT ?= $(LOCALBIN)/gofumpt
SWAG ?= $(LOCALBIN)/swag
GOLINTER ?=$(LOCALBIN)/golangci-lint
GOIMPORTS ?=$(LOCALBIN)/goimports-reviser




CLUSTER 	?= metadata-reflector
GO 			?= go
KIND 		?= kind
DOCKER 		?= docker
DOCKER_ARGS ?= --load
APP_NAME 	?= ghcr.io/nccloud/metadata-reflector
TAG 		?= 0.1.0-dev
IMG 		?= ${APP_NAME}:${TAG}
KIND_IMAGE 	?= kindest/node:v1.31.0

GOLANGCILINT_IMAGE=docker.io/golangci/golangci-lint:v1.62.0


.PHONY: help
help: ## Show help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build binary.
	$(GO) build -o bin/metadata-reflector cmd/manager/main.go

.PHONY: docker-build
docker-build: ## Build docker image.
	$(DOCKER) buildx build -t $(IMG) . $(DOCKER_ARGS)

.PHONY: docker-load
docker-load: ## Load docker image in KIND.
	$(KIND) load docker-image --name $(CLUSTER) $(IMG)

.PHONY: cluster
cluster: ## Create a single node kind cluster.
	$(KIND) create cluster --name $(CLUSTER) --image $(KIND_IMAGE)

.PHONY: cluster-delete
cluster-delete: ## Delete the kind cluster.
	$(KIND) delete cluster --name $(CLUSTER)

.PHONY: lint
lint: ## Run linter.
	$(DOCKER) run -t --rm -v $(PWD):/app -w /app $(GOLANGCILINT_IMAGE) golangci-lint run -v --timeout=10m

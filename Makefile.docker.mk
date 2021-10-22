SHELL := bash

IMAGE_NAME ?= blacklister
DOCKER_REGISTRY ?= localhost:5000
BUILD_TARGETS ?= dbmigrations prod
VERSION ?= $(shell cat VERSION)

docker:								  ## Build and push all docker images
docker: docker-build docker-push

docker-build:						## Build all docker images
docker-build: $(addprefix docker-build/,$(BUILD_TARGETS))

docker-build/%: docker-lint
	@echo "Building Docker Image $*..."; \
		docker build -t "$(DOCKER_REGISTRY)/$(IMAGE_NAME)-$*:$(VERSION)" -f "Dockerfile" --target "$*" .

docker-push: docker-build $(addprefix docker-push/,$(BUILD_TARGETS))

docker-push/%:
	@echo "Pushing Docker Image $*..."; \
	docker push "$(DOCKER_REGISTRY)/$(IMAGE_NAME)-$*:$$(cat VERSION)"

docker-lint:
	@echo "Running Dockerfile linter..."; \
	docker run --rm -i hadolint/hadolint < "Dockerfile" &&\
		echo OK


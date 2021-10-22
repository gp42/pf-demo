SHELL := bash

BUILD := $(shell git rev-parse HEAD)
VERSION := $(shell cat VERSION)
LDFLAGS = -ldflags "-X=main.gitCommit=$(BUILD) -X=main.version=$(VERSION)"
ALL_FILES := $(shell find . -type f -name '*.go') VERSION Makefile

BINARIES := $(patsubst cmd/%.go,%,$(wildcard cmd/*.go))

RELEASE ?= dev-blacklister
NAMESPACE ?= $(RELEASE)

PGDATABASE ?= blacklister
# This is not a standard postgres variable but needed to get correct secrets in migration jobs
PG_SCHEMA ?= blacklister

# Working with local development cluster (true) or not (false)
LOCAL_DEV ?= true

.PHONY: all
all:												 ## Set up local cluster, docker registry, build and deploy
	@echo "Setting up local cluster..."
	$(MAKE) local-cluster
	@echo "Building docker images..."
	$(MAKE) docker
	@echo "Deploying..."
	$(MAKE) deploy

.PHONY: all
delete-all:									 ## Delete deployment and local cluster
	@echo "Uninstalling Helm Release..."
	$(MAKE) delete-deploy || true
	@echo "Removing local cluster..."
	$(MAKE) delete-local-cluster
	@echo "Cleaning build directory..."
	$(MAKE) clean

# Include other Makefiles
include *.mk

godoc:									     ## Run GoDoc server
	@echo -e "Starting documentation server...\n\nAccess it at: http://localhost:6060/pkg/github.com/gp42/pf-demo/\n"; \
	godoc -http=:6060

.PHONY: build-all
build-all: $(addprefix build/,$(BINARIES))

.PRECIOUS: build/%-darwin-amd64 build/%-linux-amd64
build/%: build/%-darwin-amd64 build/%-linux-amd64
	@echo Done building.

build/%-darwin-amd64: build test $(ALL_FILES)
	GOOS=darwin GOARCH=amd64 go build -o $@ $(LDFLAGS) cmd/$*.go

build/%-linux-amd64: build test $(ALL_FILES)
	GOOS=linux GOARCH=amd64 go build -o $@ $(LDFLAGS) cmd/$*.go

build:
	mkdir ./build

lint:
	go vet ./...
	go fmt ./...

.PHONY: test
test: lint
	go test -v ./...

.PHONY: clean
clean:
	rm -rf ./build

.PHONY: deploy
deploy:					 		         ## Deploy Helm Release
	helm upgrade \
		--kube-context "$(K8S_CONTEXT)" \
		--install \
		-n "$(NAMESPACE)" \
		--create-namespace \
		"$(RELEASE)" \
		k8s/charts/blacklister \
		--set nodePortForwarderEnabled='$(LOCAL_DEV)' \
		--set image.tag='$(VERSION)' \
		--set image.registry='$(DOCKER_REGISTRY)'

.PHONY: delete-deploy
delete-deploy:						   ## Delete Helm Release
	helm uninstall \
		--kube-context "$(K8S_CONTEXT)" \
	 "$(RELEASE)"

.PHONY: migrate-up
migrate:
	export PGUSER="$$(kubectl get secret --context "$(K8S_CONTEXT)" -n "$(NAMESPACE)" "$(PGDATABASE)-$(PG_SCHEMA)-owner-user.ops-$(RELEASE)-db.credentials.postgresql.acid.zalan.do" -o go-template='{{.data.username|base64decode}}')" &&\
	export PGPASSWORD="$$(kubectl get secret  --context "$(K8S_CONTEXT)" -n "$(NAMESPACE)" "$(PGDATABASE)-$(PG_SCHEMA)-owner-user.ops-$(RELEASE)-db.credentials.postgresql.acid.zalan.do" -o go-template='{{.data.password|base64decode}}')" &&\
		migrate -path=db/migrations/ -database postgres://localhost:5432/$(PGDATABASE)?sslmode=require $(CMD)

.PHONY: migrate-up
migrate-up: CMD = up
migrate-up: migrate				   ## Run DB Migrations

.PHONY: migrate-down
migrate-down: CMD = down 1 
migrate-down: migrate			   ## Rollback last DB Migration

.PHONY: migrate-drop
migrate-drop: CMD = drop -f
migrate-drop: migrate			   ## Drop DB

psql:											   ## Login into the database with psql as Owner
	export PGUSER="$$(kubectl get secret --context "$(K8S_CONTEXT)" -n "$(NAMESPACE)" "$(PGDATABASE)-$(PG_SCHEMA)-owner-user.ops-$(RELEASE)-db.credentials.postgresql.acid.zalan.do" -o go-template='{{.data.username|base64decode}}')" &&\
	export PGPASSWORD="$$(kubectl get secret  --context "$(K8S_CONTEXT)" -n "$(NAMESPACE)" "$(PGDATABASE)-$(PG_SCHEMA)-owner-user.ops-$(RELEASE)-db.credentials.postgresql.acid.zalan.do" -o go-template='{{.data.password|base64decode}}')" &&\
		psql -h localhost -d $(PGDATABASE)

help:                    	   ## Show help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[$$()% 0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

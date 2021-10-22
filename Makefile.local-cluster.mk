SHELL := bash

CLUSTER_NAME ?= $(shell cat k8s/local/kind.yaml | grep '^name:' | awk '{print $$2}')
REGISTRY_NAME ?= kind-registry
KIND_CONFIG ?= k8s/local/kind.yaml
KIND_IMAGE ?= kindest/node:v1.21.1@sha256:69860bda5563ac81e3c0057d654b5253219618a22ec3a346306239bba8cfa1a6
CLUSTER_REGISTRY_CONFIG ?= k8s/local/registry-cm.yaml

PG_OPERATOR_LINK ?= https://github.com/zalando/postgres-operator/raw/master/charts/postgres-operator/postgres-operator-1.7.0.tgz
PG_OPERATOR_NS ?= postgres-operator-system

K8S_CONTEXT ?= kind-$(CLUSTER_NAME)

.PHONY: local-cluster
local-cluster:               ## Set up Kubernetes cluster with all dependencies
local-cluster: kind docker-registry config-local-cluster postgres-operator

.PHONY: delete-local-cluster
delete-local-cluster:        ## Delete Kubernetes cluster and its dependencies
delete-local-cluster: delete-kind delete-docker-registry

.PHONY: local-cluster-config
config-local-cluster:
	@echo "Configuring cluster...";\
	{ docker network connect "kind" "$(REGISTRY_NAME)" || true; } &&\
		kubectl --context "$(K8S_CONTEXT)" apply -f "$(CLUSTER_REGISTRY_CONFIG)" &&\
			echo OK

.PHONY: kind
kind:
	kind create cluster --config "$(KIND_CONFIG)" --image "$(KIND_IMAGE)"

.PHONY: delete-kind
delete-kind:
	kind delete cluster --name "$(CLUSTER_NAME)"
	rm -rfv ./var

.PHONY: docker-registry
docker-registry:
	@echo "Creating Docker Registry..."; \
	running="$$(docker inspect -f '{{.State.Running}}' "$(REGISTRY_NAME)" 2>/dev/null || true)" &&\
		if [ "$$running" != 'true' ]; then \
			docker run \
				-d --restart=always -p "127.0.0.1:5000:5000" --name "$(REGISTRY_NAME)" \
				registry:2 &&\
					echo OK; \
		fi

.PHONY: delete-docker-registry
delete-docker-registry:
	@echo "Deleting Docker Registry..."; \
	docker stop "$(REGISTRY_NAME)" &&\
		docker rm "$(REGISTRY_NAME)" &&\
			echo OK || true

.PHONY: postgres-operator
postgres-operator: var/postgres-operator.tgz config-local-cluster local-cluster
	helm install \
		--namespace "$(PG_OPERATOR_NS)" \
		--create-namespace \
		postgres-operator ./var/postgres-operator.tgz &&\
			kubectl rollout status --context "$(K8S_CONTEXT)" -n "$(PG_OPERATOR_NS)" -w deploy/postgres-operator

var/postgres-operator.tgz:
	@mkdir -p var &&\
		wget -L "$(PG_OPERATOR_LINK)" \
			-O var/postgres-operator.tgz 

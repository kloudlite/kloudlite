IMAGE_REGISTRY := "harbor.dev.madhouselabs.io/kloudlite/hotspot/api"
TAG :="go"
DOCKERFILE := "./Dockerfile"
CMD_ARGS := ""
# SHELL := $(shell which bash)

BUILD_ARGS := "--build-arg APP=$(APP) --build-arg CMD_ARGS=$(CMD_ARGS)"

.docker: IMAGE_NAME=$(shell echo $(APP) | tr "/" "-" | sed 's/\-/\//')
.docker: .build .push

# .build: BUILD_ID=$(shell date -Iseconds)
.build:
	cd $(shell dirname ${DOCKERFILE}) && eval docker build -f $(shell basename ${DOCKERFILE}) -t $(IMAGE_REGISTRY)/${IMAGE_NAME}:${TAG} . \
		$(BUILD_ARGS)

.push:
	docker push $(IMAGE_REGISTRY)/${IMAGE_NAME}:${TAG}

image.$(APP): .docker

run.$(APP):
	$(eval include apps/$(APP)/.env)
	$(eval export)
	@go mod tidy
	go run apps/$(APP)/main.go $(CMD_ARGS)

build.$(APP):
	@go mod tidy
	go build -o bin/$(APP) apps/$(APP)/main.go

start.$(APP):
	./bin/$(APP) $(CMD_ARGS)

.runner:
	@make run.$(APP) -e APP=$(APP) -e CMD_ARGS="--dev"
	
dev:
	make dev.producer
	make dev.consumer


dev.producer: APP=message-producer
dev.producer: .runner

dev.consumer: APP=message-consumer
dev.consumer: .runner

image.consumer: APP=message-consumer
image.consumer:
	@make image.$(APP) -e APP=$(APP)

dev.wireguard: APP=wireguard
dev.wireguard: .runner

wireguard.dev: APP=wireguard
wireguard.dev:
	cd apps/wireguard/internal/app && go run github.com/99designs/gqlgen
	@make .runner -e APP=$(APP)

wireguard.gql:
	cd apps/wireguard/internal/app && go run github.com/99designs/gqlgen

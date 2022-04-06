.PHONY: start dev build docker

APP := "message-consumer"

start:
	./bin/$(APP)

dev:
	go run ./main.go --dev

build:
	go mod tidy
	go build -tags musl -o bin/$(APP) ./main.go

docker: IMAGE_REGISTRY = "harbor.dev.madhouselabs.io/kloudlite/hotspot/api"
docker: TAG = "go"
docker: IMAGE_NAME = $(APP)
docker: _dkr.depends_on _dkr.build _dkr.push

_dkr.depends_on: PROJECT_ROOT := "../.."
_dkr.depends_on:
	cd $(PROJECT_ROOT) && DOCKER_BUILDKIT=1 docker build . -t shared-lib:api-go

_dkr.build: PATH_TO_APP := "./apps/$(APP)"
_dkr.build:
	DOCKER_BUILDKIT=1 docker build . -t $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(TAG) --build-arg APP_DIR=$(PATH_TO_APP)

_dkr.push:
	docker push $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(TAG)

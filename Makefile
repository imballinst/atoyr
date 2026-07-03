DOCKER_IMAGE ?= atoyr
DOCKER_TAG ?= latest

.PHONY: build up down build-server build-client clean

build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

build-server:
	docker build --target server-build -t atoyr-server:latest .

build-client:
	docker build --target client-build -t atoyr-client:latest .

clean:
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true

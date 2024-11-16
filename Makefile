SHELL := /bin/bash
PWD := $(shell pwd)

all:

build:
	docker build -f ./build/gateway.Dockerfile -t "gateway:latest" .
	docker build -f ./build/review.Dockerfile -t "reviews-filter:latest" .
.PHONY: build

build-client:
	docker build -f ./build/client.Dockerfile -t "client:latest" .
.PHONY: build-client

build-hc:
	docker build -f ./build/healthchecker.Dockerfile -t "healthcheck:latest" .
.PHONY: build-hc

docker-compose-up: build
	docker compose -f docker-compose.yaml up -d --build
.PHONY: docker-compose-up

docker-compose-up-client: build-client
	docker compose -f docker-compose-client.yaml up -d --build
.PHONY: docker-compose-client

docker-compose-up-hc: build-hc
	docker compose -f docker-compose-hc.yaml up -d --build
.PHONY: docker-compose-up-hc


docker-compose-down:
	docker compose -f docker-compose.yaml stop
	docker compose -f docker-compose.yaml down
.PHONY: docker-compose-down

docker-compose-down-client:
	docker compose -f docker-compose-client.yaml stop
	docker compose -f docker-compose-client.yaml down
.PHONY: docker-compose-down-client

docker-compose-down-hc:
	docker compose -f docker-compose-hc.yaml stop
	docker compose -f docker-compose-hc.yaml down
.PHONY: docker-compose-down-hc

docker-compose-down-all:
	docker compose -f docker-compose-client.yaml stop
	docker compose -f docker-compose-client.yaml down
	docker compose -f docker-compose.yaml stop
	docker compose -f docker-compose.yaml down
	docker compose -f docker-compose-hc.yaml stop
	docker compose -f docker-compose-hc.yaml down
.PHONY: docker-compose-down-all

docker-compose-logs:
	docker compose -f docker-compose.yaml logs -f
.PHONY: docker-compose-logs

docker-compose-logs-client:
	docker compose -f docker-compose-client.yaml logs -f
.PHONY: docker-compose-logs-client

docker-compose-logs-hc:
	docker compose -f docker-compose-hc.yaml logs -f
.PHONY: docker-compose-logs-hc
SHELL := /bin/bash
PWD := $(shell pwd)

all:

build:
	docker build -f ./build/joiner_percentile.Dockerfile -t "joiner-percentile-filter:latest" .
	docker build -f ./build/joiner_counter.Dockerfile -t "joiner-counter-filter:latest" .
	docker build -f ./build/joiner_top.Dockerfile -t "joiner-top-filter:latest" .
	docker build -f ./build/review.Dockerfile -t "reviews-filter:latest" .
	docker build -f ./build/text.Dockerfile -t "review-text-filter:latest" .
	docker build -f ./build/action.Dockerfile -t "action-filter:latest" .
	docker build -f ./build/indie.Dockerfile -t "indie-filter:latest" .
	docker build -f ./build/platform.Dockerfile -t "platform-filter:latest" .
	docker build -f ./build/gateway.Dockerfile -t "gateway:latest" .
	docker build -f ./build/client.Dockerfile -t "client:latest" .
	docker build -f ./build/topn.Dockerfile -t "topn:latest" .
.PHONY: build

build-client:
	docker build -f ./build/client.Dockerfile -t "client:latest" .
.PHONY: build-client

docker-compose-up: build
	docker compose -f docker-compose.yaml up -d --build
.PHONY: docker-compose-up

docker-compose-up-client: build-client
	docker compose -f docker-compose-client.yaml up -d --build
.PHONY: docker-compose-client

docker-compose-down:
	docker compose -f docker-compose.yaml stop
	docker compose -f docker-compose.yaml down
.PHONY: docker-compose-down

docker-compose-down-client:
	docker compose -f docker-compose-client.yaml stop -t 1
	docker compose -f docker-compose-client.yaml down
.PHONY: docker-compose-down-client

docker-compose-down-all:
	docker compose -f docker-compose-client.yaml stop -t 1
	docker compose -f docker-compose-client.yaml down
	docker compose -f docker-compose.yaml stop -t 1
	docker compose -f docker-compose.yaml down
.PHONY: docker-compose-down-all

docker-compose-logs:
	docker compose -f docker-compose.yaml logs -f
.PHONY: docker-compose-logs

docker-compose-logs-client:
	docker compose -f docker-compose-client.yaml logs -f
.PHONY: docker-compose-logs-client

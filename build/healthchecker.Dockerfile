FROM golang:latest

WORKDIR /app

# Install Docker
RUN apt-get update && apt-get install -y \
    curl \
    && curl -fsSL https://get.docker.com | sh

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/healthchecker/healthchecker.go ./main.go
COPY pkg/ ./pkg/
COPY internal/healthcheck ./internal/healthcheck/

RUN CGO_ENABLED=0 GOOS=linux go build -o /main

ENTRYPOINT ["/main"]
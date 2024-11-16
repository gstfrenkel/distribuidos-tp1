FROM golang:latest

WORKDIR /app

# Install Docker
RUN apt-get update && apt-get install -y \
    curl \
    && curl -fsSL https://get.docker.com | sh

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/health_checker/health_checker.go ./main.go
COPY pkg/ ./pkg/
COPY internal/health_check ./internal/health_check/

RUN CGO_ENABLED=0 GOOS=linux go build -o /main

ENTRYPOINT ["/main"]
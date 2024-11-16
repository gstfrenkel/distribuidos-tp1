FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN apt-get update && apt-get install -y netcat-openbsd

# Update path to desired entrypoint
COPY cmd/gateway/gateway.go ./main.go
COPY pkg/ ./pkg/
COPY internal/gateway ./internal/gateway/
COPY internal/healthcheck ./internal/healthcheck/

RUN CGO_ENABLED=0 GOOS=linux go build -o /main

ENTRYPOINT ["/main"]
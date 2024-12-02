FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Update path to desired entrypoint
COPY cmd/worker/joiner/top.go ./main.go
COPY pkg/ ./pkg/
COPY internal/errors/ ./internal/errors/
COPY internal/worker/worker.go ./internal/worker/
COPY internal/worker/joiner/joiner.go ./internal/worker/joiner/
COPY internal/worker/joiner/top.go ./internal/worker/joiner/
COPY internal/healthcheck/service.go ./internal/healthcheck/service.go

RUN CGO_ENABLED=0 GOOS=linux go build -o /main

ENTRYPOINT ["/main"]
FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Update path to desired entrypoint
COPY cmd/worker/percentile/percentile.go ./main.go
COPY pkg/ ./pkg/
COPY internal/errors/ ./internal/errors/
COPY internal/worker/worker.go ./internal/worker/
COPY internal/worker/percentile/percentile.go ./internal/worker/percentile/
COPY internal/worker/percentile/batch_utils.go ./internal/worker/percentile/
COPY internal/healthcheck/healthcheck_service.go ./internal/healthcheck/healthcheck_service.go

RUN CGO_ENABLED=0 GOOS=linux go build -o /main

ENTRYPOINT ["/main"]
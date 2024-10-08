FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Update path to desired entrypoint
COPY cmd/worker/joiner_percentile/percentile.go ./main.go
COPY pkg/ ./pkg/
COPY internal/errors/ ./internal/errors/
COPY internal/worker/worker.go ./internal/worker/
# Update path to desired entrypoint
COPY internal/worker/joiner/percentile.go ./internal/worker/joiner/

# Replace with volume
COPY configs/joiner_percentile.json config.json

ENTRYPOINT ["go", "run", "main.go"]
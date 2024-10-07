FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Update path to desired entrypoint
COPY cmd/worker/platform_counter_agg/platform_counter_agg.go ./main.go
COPY pkg/ ./pkg/
COPY internal/errors/ ./internal/errors/
COPY internal/worker/worker.go ./internal/worker/
# Update path to desired entrypoint
COPY internal/worker/platform_counter_agg/platform_counter_agg.go ./internal/worker/platform_counter_agg/

# Replace with volume
COPY configs/platform_counter_agg.json config.json

ENTRYPOINT ["go", "run", "main.go"]
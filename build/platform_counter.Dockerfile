FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Update path to desired entrypoint
COPY cmd/worker/platform_counter/platform_counter.go ./main.go
COPY pkg/ ./pkg/
COPY internal/errors/ ./internal/errors/
COPY internal/worker/worker.go ./internal/worker/
# Update path to desired entrypoint
COPY internal/worker/platform_counter/platform_counter.go ./internal/worker/platform_counter/

# Replace with volume
COPY configs/platform_counter.json config.json

ENTRYPOINT ["go", "run", "main.go"]
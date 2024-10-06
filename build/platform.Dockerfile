FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Replace with volume
COPY configs/platform.json config.json

# Update path to desired entrypoint
COPY cmd/worker/platform/platform.go ./main.go
COPY pkg/ ./pkg/
COPY internal/errors/ ./internal/errors/
COPY internal/worker/worker.go ./internal/worker/
# Update path to desired entrypoint
COPY internal/worker/platform/platform.go ./internal/worker/platform/

ENTRYPOINT ["go", "run", "main.go"]
FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Replace with volume
COPY configs/indie.json config.json

# Update path to desired entrypoint
COPY cmd/worker/indie/indie.go ./main.go
COPY pkg/ ./pkg/
COPY internal/errors/ ./internal/errors/
COPY internal/worker/worker.go ./internal/worker/
# Update path to desired entrypoint
COPY internal/worker/indie/indie.go ./internal/worker/indie/

ENTRYPOINT ["go", "run", "main.go"]
FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Update path to desired entrypoint
COPY cmd/worker/release_date/release_date.go ./main.go
COPY pkg/ ./pkg/
COPY internal/errors/ ./internal/errors/
COPY internal/worker/worker.go ./internal/worker/
# Update path to desired entrypoint
COPY internal/worker/release_date/release_date.go ./internal/worker/release_date/

# Replace with volume
COPY configs/release_date.json config.json

ENTRYPOINT ["go", "run", "main.go"]
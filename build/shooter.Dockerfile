FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Replace with volume
COPY configs/shooter.toml config.toml

# Update path to desired entrypoint
COPY cmd/worker/shooter/shooter.go ./main.go
COPY pkg/ ./pkg/
COPY internal/ ./internal/

ENTRYPOINT ["go", "run", "main.go"]
FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Replace with volume
COPY configs/topn.json config.json

# Update path to desired entrypoint
COPY cmd/worker/top_n/top_n.go ./main.go
COPY pkg/ ./pkg/
COPY internal/ ./internal/

ENTRYPOINT ["go", "run", "main.go"]
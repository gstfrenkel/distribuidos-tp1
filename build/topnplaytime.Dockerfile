FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Update path to desired entrypoint
COPY cmd/worker/top_n_playtime/top_n_playtime.go ./main.go
COPY pkg/ ./pkg/
COPY internal/errors/ ./internal/errors/
COPY internal/worker/worker.go ./internal/worker/
# Update path to desired entrypoint
COPY internal/worker/top_n_playtime/top_n_playtime.go ./internal/worker/top_n_playtime/

# Replace with volume
COPY configs/topn_playtime.json config.json

ENTRYPOINT ["go", "run", "main.go"]
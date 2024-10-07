FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Update path to desired entrypoint
COPY cmd/worker/top_n_playtime_agg/top_n_playtime_agg.go ./main.go
COPY pkg/ ./pkg/
COPY internal/errors/ ./internal/errors/
COPY internal/worker/worker.go ./internal/worker/
# Update path to desired entrypoint
COPY internal/worker/top_n_playtime_agg/top_n_playtime_agg.go ./internal/worker/top_n_playtime_agg/

# Replace with volume
COPY configs/top_n_playtime_agg.json config.json

ENTRYPOINT ["go", "run", "main.go"]
FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Update path to desired entrypoint
COPY cmd/worker/top_n/top_n.go ./main.go
COPY pkg/ ./pkg/
COPY internal/errors/ ./internal/errors/
COPY internal/worker/worker.go ./internal/worker/
# Update path to desired entrypoint
COPY internal/worker/top_n/top_n.go ./internal/worker/top_n/

# Replace with volume todo valen CUANDO HAGAMOS ESO, BORRAR ESTE DOCKERFILE
COPY configs/topn_agg.json config.json

ENTRYPOINT ["go", "run", "main.go"]
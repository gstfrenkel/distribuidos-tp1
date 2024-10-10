FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Update path to desired entrypoint
COPY cmd/worker/joiner_percentile/percentile.go ./main.go
COPY pkg/ ./pkg/
COPY internal/errors/ ./internal/errors/
COPY internal/worker/worker.go ./internal/worker/
# Update path to desired entrypoint
COPY internal/worker/joiner/percentile.go ./internal/worker/joiner/

RUN CGO_ENABLED=0 GOOS=linux go build -o /main

ENTRYPOINT ["/main"]
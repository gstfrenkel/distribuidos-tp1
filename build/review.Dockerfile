FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN apt-get update && apt-get install -y netcat-openbsd

# Update path to desired entrypoint
COPY cmd/worker/filter/review.go ./main.go
COPY pkg/ ./pkg/
COPY internal/errors/ ./internal/errors/
COPY internal/worker/worker.go ./internal/worker/
COPY internal/worker/filter/review.go ./internal/worker/filter/
COPY internal/healthcheck/service.go ./internal/healthcheck/service.go

RUN CGO_ENABLED=0 GOOS=linux go build -o /main

ENTRYPOINT ["/main"]
FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN apt-get update && apt-get install -y netcat-openbsd

# Update path to desired entrypoint
COPY cmd/worker/review/review.go ./main.go
COPY pkg/ ./pkg/
COPY internal/errors/ ./internal/errors/
COPY internal/worker/worker.go ./internal/worker/
COPY internal/worker/review/review.go ./internal/worker/review/
COPY internal/health_check/health_check_service.go ./internal/health_check/health_check_service.go

RUN CGO_ENABLED=0 GOOS=linux go build -o /main

ENTRYPOINT ["/main"]
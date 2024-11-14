FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/healthchecker/health_checker.go ./main.go
COPY pkg/ ./pkg/
COPY internal/healthcheck ./internal/healthcheck/

RUN CGO_ENABLED=0 GOOS=linux go build -o /main

ENTRYPOINT ["/main"]
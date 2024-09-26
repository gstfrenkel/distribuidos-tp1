FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY cmd/workers/review_filter/review_filter.go ./main.go

ENTRYPOINT ["go", "run", "main.go"]
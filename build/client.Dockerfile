FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Update path to desired entrypoint
COPY cmd/client/client.go ./main.go
COPY pkg/ ./pkg/
COPY internal/client/client.go ./internal/client/

ENTRYPOINT ["go", "run", "main.go"]
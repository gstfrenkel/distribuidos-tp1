FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Update path to desired entrypoint
COPY cmd/gateway/gateway.go ./main.go
COPY pkg/ ./pkg/
COPY internal/ ./internal/

# Compilar la aplicación Go
RUN go build -o /app/gateway main.go

ENTRYPOINT ["/app/gateway"]
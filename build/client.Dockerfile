FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Update path to desired entrypoint
COPY cmd/client/client.go ./main.go
COPY pkg/ ./pkg/
COPY internal/ ./internal/

RUN go build -o /app/client main.go

# Compilar la aplicación Go
RUN go build -o /app/client main.go

ENTRYPOINT ["/app/client"]
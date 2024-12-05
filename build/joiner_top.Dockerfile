FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Update path to desired entrypoint
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./cmd/worker/joiner/top.go

FROM scratch

# Copy the compiled binary from the builder stage
COPY --from=builder /main /main

ENTRYPOINT ["/main"]
FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN apt-get update && apt-get install -y netcat-openbsd

# Update path to desired entrypoint
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./cmd/worker/filter/review.go

ENTRYPOINT ["/main"]

FROM scratch

# Copy the compiled binary from the builder stage
COPY --from=builder /main /main

ENTRYPOINT ["/main"]
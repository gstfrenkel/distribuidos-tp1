FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Update path to desired entrypoint
COPY cmd/worker/review/review.go ./main.go
COPY internal/ pkg/ ./

ENTRYPOINT ["go", "run", "main.go"]
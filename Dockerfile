# syntax=docker/dockerfile:1

FROM golang:1.24.5-alpine AS builder

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Adjust path based on your project layout (assumes cmd/server/main.go exists)
WORKDIR /app/cmd/server
RUN go build -o /go/bin/main

# Runtime stage
FROM alpine:latest

COPY --from=builder /go/bin/main /main

EXPOSE 8080

CMD ["/main"]
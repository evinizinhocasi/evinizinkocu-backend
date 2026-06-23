# Multi-stage Dockerfile for Go Backend
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install git/certs
RUN apk add --no-cache git ca-certificates

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and embed files
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api

# Run stage
FROM alpine:3.18

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/api /app/api

EXPOSE 8080

CMD ["/app/api"]

# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Copying .env; seek out better alternative
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o api ./cmd/api

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Optional: install CA certificates for HTTPS requests
# RUN apk --no-cache add ca-certificates

COPY --from=builder /app/api .

EXPOSE 8000

CMD [ "./api" ]
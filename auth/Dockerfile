FROM golang:1.24-alpine AS builder

WORKDIR /cmd
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o app ./cmd/auth/
FROM alpine:3.21
WORKDIR /cmd
COPY --from=builder /cmd/app .
RUN mkdir -p /config
COPY --from=builder /cmd/config/ ./config/
COPY --from=builder /cmd/migrations/ ./migrations/
CMD ["./app"]
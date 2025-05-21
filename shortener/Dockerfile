FROM golang:1.23-alpine AS builder

WORKDIR /cmd
COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/swaggo/swag/cmd/swag@latest
COPY . .
RUN swag init -g ./cmd/linkify/main.go -o ./docs
RUN CGO_ENABLED=0 go build -o app ./cmd/linkify/

FROM alpine:3.21
WORKDIR /cmd
COPY --from=builder /cmd/app .
RUN mkdir -p /config
COPY --from=builder /cmd/config/prod.yaml ./config/
COPY --from=builder /cmd/docs .
CMD ["./app"]
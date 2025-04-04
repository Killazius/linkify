FROM golang:1.23-alpine AS builder

WORKDIR /cmd
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o app ./cmd/linkify/


FROM alpine:3.21
WORKDIR /cmd
COPY --from=builder /cmd/app .
COPY --from=builder /cmd/.env .
COPY --from=builder /cmd/config/prod.yaml .
COPY --from=builder /cmd/docs .
CMD ["./app"]
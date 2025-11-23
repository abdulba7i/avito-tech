FROM golang:1.25-alpine AS builder

WORKDIR "/app"

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy

RUN go build -o app ./cmd/service/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/app .
COPY --from=builder /app/migrations ./migrations

CMD ["./app"]
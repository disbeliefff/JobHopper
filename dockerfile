
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /job_hunter_bot ./cmd/job_hunter_bot

FROM alpine:3.18

RUN apk add --no-cache postgresql-client

COPY --from=builder /job_hunter_bot /usr/local/bin/job_hunter_bot
COPY --from=builder /app/internal/storage/migrations /migrations

WORKDIR /usr/local/bin

CMD ["job_hunter_bot"]

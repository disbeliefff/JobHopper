FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/vacancy-hunter ./cmd/

EXPOSE 8080
CMD ["/app/vacancy-hunter"]

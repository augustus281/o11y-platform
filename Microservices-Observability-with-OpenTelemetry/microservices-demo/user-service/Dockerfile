FROM golang:1.25.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o user-service cmd/main.go

FROM alpine:latest

COPY --from=builder /app/user-service /usr/local/bin/user-service

EXPOSE 8888

CMD ["user-service"]
FROM golang:1.20 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o j2lab

FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/j2lab /usr/local/bin/j2lab
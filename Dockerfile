FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o nimbuslb ./cmd/server

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/nimbuslb .
COPY configs ./configs

EXPOSE 8080

ENV NIMBUSLB_CONFIG=configs/config.docker.yaml

ENTRYPOINT ["./nimbuslb"]
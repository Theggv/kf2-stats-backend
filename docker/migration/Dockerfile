FROM golang:alpine AS builder

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

RUN apk update && \
    apk add --no-cache git && \
    apk add --update gcc musl-dev

RUN mkdir /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/main -a -ldflags '-linkmode external -extldflags "-static"' ./cmd/migration

FROM scratch

COPY --from=builder /app/main /app/main

ENTRYPOINT [ "/app/main" ]
FROM golang:1.24.1-alpine AS builder

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

RUN apk update && \
    apk add --no-cache git openssl ca-certificates && \
    apk add --update gcc musl-dev && \
    openssl req -newkey rsa:2048 -nodes -keyout key -x509 \
    -sha256 -days 3650 -subj /CN=localhost -out crt && \
    update-ca-certificates

RUN mkdir /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/main -a -ldflags '-linkmode external -extldflags "-static"' ./cmd/production

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/main /app/main
EXPOSE 3000

ENTRYPOINT [ "/app/main" ]
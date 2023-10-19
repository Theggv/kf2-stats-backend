FROM golang:alpine AS builder

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

RUN apk update \
    && apk add --no-cache git \
    && apk add --update gcc musl-dev

RUN mkdir /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init
RUN go build -o /app/main -a -ldflags '-linkmode external -extldflags "-static"' .

FROM scratch

COPY --from=builder /app/main /go/bin/kf2-stats-backend
EXPOSE 3000

ENTRYPOINT [ "/go/bin/kf2-stats-backend" ]
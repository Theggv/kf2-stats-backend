## Requirements

- MySQL 8.0+ (non-docker setup)
- Go 1.21+

## Setup

- Create `.env` based on `.env.example`
- Fill MySQL variables.
- Set `SECRET_TOKEN` as random string. Used to protect POST endpoints called from the mutator.
- Set `STEAM_API_KEY` from https://steamcommunity.com/dev/apikey in order to show user avatars in frontend.

## Development

### Init
```
go install github.com/cosmtrek/air@latest
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g .\cmd\development\main.go
```
### Run
```
air
```

### Swagger Docs
Generated docs are available on http://localhost:3000/docs/index.html
- You have to rerun `swag init -g .\cmd\development\main.go` every time when api changes

## Build

```
go build -o ./main.exe -a -ldflags '-linkmode external -extldflags "-static"' ./cmd/production
```

### Docker
```
docker compose up -d --build
```
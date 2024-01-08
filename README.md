# Better Stats Backend

Backend part of Better Stats for Killing Floor 2.

## Non-docker setup

### Requirements

- MySQL 8.0+
- Go 1.21+

### Env

- Create `.env` based on `.env.example`
- Use `ip:port` format for `SERVER_ADDR` (default port is 3000)
- Fill MySQL variables.
- Set `SECRET_TOKEN` as random string. Used to protect POST endpoints called from the mutator.
- Set `STEAM_API_KEY` from https://steamcommunity.com/dev/apikey. Used to show user avatars on frontend.

### Production build

```
go build -o ./main.exe -a -ldflags '-linkmode external -extldflags "-static"' ./cmd/production
```

### Development build

- Install air for live reload and swag to generate docs

```
go install github.com/cosmtrek/air@latest
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g .\cmd\development\main.go
```

- Use air

```
air
```

**NOTE:**
Generated docs are available on http://localhost:3000/docs/index.html. You have to rerun `swag init -g .\cmd\development\main.go` every time when api changes in order to see actual data.

## Docker setup

1. Complete [Env](#env) step from Non-docker setup.
2. Set `SERVER_ADDR` to `0.0.0.0:3000`
3. Build using docker compose

```
docker compose up -d --build
```

4. If changes were made run:

```
docker compose down
docker compose up -d --build
```


# Auth Service

Minimal Go auth service skeleton using Chi, Postgres (pgx), goose migrations, and dotenv configuration.

## Environment Variables

Use `.env.Dev` for local development overrides and a generic `.env` for shared defaults. Required variables:

```
MODE=development            # development | production
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=authuser
DATABASE_PASSWORD=DevPass123!
DATABASE_NAME=auth_db
PORT=8082                   # HTTP listen port
```

`.env.Dev` is loaded first when `MODE=development`, then `.env` as a fallback.

## Migrations
In development and test modes migrations run automatically at startup. For manual control:

```
go run github.com/pressly/goose/v3/cmd/goose@latest create add_sessions_table sql
go run github.com/pressly/goose/v3/cmd/goose@latest -dir ./migrations postgres "$(MODE=development ./print_dsn.sh)" up
```

(You can write a small helper script that loads env and prints the DSN using the same logic as `cfg.DSN()`.)

## Running

```
MODE=development go run ./cmd/auth-service
```

Health check:
```
curl localhost:8082/api/health
```

## Notes / Next Steps
- Add auth endpoints in `internal/api/router.go`.
- Consider middleware for logging and request IDs.
- Add integration tests hitting a test database (spin up Postgres via container or local instance).

# Auth Service

## Architecture
- `internal/api`: router setup, middleware, request structs, and HTTP handlers.
- `internal/service`: registration, login, token rotation, logout, and cookie handling.
- `internal/repository`: Postgres access for users, password credentials, and refresh tokens.
- `internal/contracts`: shared domain models and DTOs.
- `internal/db`: database connection and migration runner.
- `internal/health`: health checks.
- `internal/config`: environment loading and DSN construction.

## Environment

Required variables are documented in `.env.example`. You must also specify a `MODE` environment variable with value
`development` or `production`.

## Migrations

Migrations run automatically. For manual runs:

```bash
go run github.com/pressly/goose/v3/cmd/goose@latest create <migration_name> sql
go run github.com/pressly/goose/v3/cmd/goose@latest -dir ./migrations postgres "<connection_string>" up
go run github.com/pressly/goose/v3/cmd/goose@latest -dir ./migrations postgres "<connection_string>" down
```

## Endpoints
Handlers return `requests.APIResponse` unless noted otherwise. The auth handlers currently build their payloads inline rather than through dedicated response structs.

- `/health`
  - `GET` - return service health.
  - `GET`, input `none`, output `requests.APIResponse`
- `/auth`
  - `/register`
    - `POST` - create a user and issue a token pair.
    - `POST`, input `RegisterRequest`, output `requests.APIResponse`
  - `/login`
    - `POST` - authenticate a user and issue a token pair.
    - `POST`, input `LoginRequest`, output `requests.APIResponse`
  - `/logout`
    - `POST` - revoke the supplied refresh token.
    - `POST`, input `LogoutRequest`, output `none` (`204 No Content`)
  - `/refresh`
    - `POST` - rotate a refresh token and return a new pair.
    - `POST`, input `RefreshRequest`, output `requests.APIResponse`

## Database
Entities:

- `users(id, username, email, created_at, updated_at, role)`
- `password_credentials(user_id, password_hash, password_salt)`
- `refresh_tokens(token_id, user_id, token_hash, issued_at, expires_at, revoked_at, replaced_by_token_id)`

Relations:

- `(password_credentials.user_id, users.id)`
- `(refresh_tokens.user_id, users.id)`
- `(refresh_tokens.replaced_by_token_id, refresh_tokens.token_id)`

## Notes
- Access tokens are short-lived JWTs.
- Refresh tokens are stored as hashes in the database.
- Register and login also set the refresh token as an HTTP cookie.
- Creating a new token pair revokes any existing active refresh tokens for that user.
- Request validation is handled in `internal/api/middleware.go`.
- Role validation applies basic normalisation before allowed-value checks.
- In development and test mode, migrations run at start-up.

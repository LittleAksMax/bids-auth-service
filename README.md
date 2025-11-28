# Auth Service

Minimal Go auth service using Chi router, Postgres (pgx), Redis cache, JWT tokens, and goose migrations.

## Environment Variables

Use `.env.Dev` for local development overrides and a generic `.env` for shared defaults. Required variables:

```
MODE=development            # development | production
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=authuser
DATABASE_PASSWORD=DevPass123!
DATABASE_NAME=auth_db
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
ACCESS_TOKEN_SECRET=your-secret-key-change-me
REFRESH_TOKEN_SECRET=your-refresh-secret-key-change-me
ACCESS_TOKEN_TTL=15m        # e.g., 15m, 1h
REFRESH_TOKEN_TTL=168h      # e.g., 168h (7 days)
PORT=8082                   # HTTP listen port
```

`.env.Dev` is loaded first when `MODE=development`, then `.env` as a fallback.

## Migrations
In development and test modes migrations run automatically at startup. For manual control:

```bash
# Create a new migration
go run github.com/pressly/goose/v3/cmd/goose@latest create <migration_name> sql

# Run migrations up
go run github.com/pressly/goose/v3/cmd/goose@latest -dir ./migrations postgres "<connection_string>" up

# Rollback migrations
go run github.com/pressly/goose/v3/cmd/goose@latest -dir ./migrations postgres "<connection_string>" down
```

## Running

```bash
MODE=development go run ./cmd/auth-service
```

## API Endpoints

### Health Check
```bash
curl http://localhost:8082/health
```

### Register a New User
```bash
curl -X POST http://localhost:8082/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "SecurePassword123!"
  }'
```

### Login
```bash
curl -X POST http://localhost:8082/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "SecurePassword123!"
  }'
```
Returns both `refresh_token` and `access_token`.

### Refresh Tokens
```bash
curl -X POST http://localhost:8082/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<your_refresh_token>"
  }'
```
Returns a new token pair (`refresh_token` and `access_token`). The old refresh token is invalidated.

### Logout
```bash
curl -X POST http://localhost:8082/auth/logout \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<your_refresh_token>"
  }'
```
Invalidates the refresh token in the cache.

### Validate Access Token (API Key Protected)
```bash
curl -X POST http://localhost:8082/tokens/validate \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: <your_validation_api_key>" \
  -d '{
    "access_token": "<jwt_access_token>"
  }'
```
Validates an access token and returns its claims (user_id, expires_at, issued_at). Requires `X-API-Key` header with the `VALIDATION_API_KEY` from environment.

### Invalidate Refresh Token (API Key Protected)
```bash
curl -X POST http://localhost:8082/tokens/invalidate \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: <your_validation_api_key>" \
  -d '{
    "refresh_token": "<refresh_token_to_invalidate>"
  }'
```
Invalidates a specific refresh token in the cache. Requires `X-Api-Key` header. Useful for admin operations or security purposes.

## Architecture

- **Router**: Chi router with routes defined in `internal/api/routes.go`
- **Controllers**: 
  - Auth endpoints in `internal/api/auth_controller.impl.go`
  - Token management endpoints in `internal/api/tokens_controller.impl.go`
- **Token Management**: JWT access tokens and random refresh tokens managed by `internal/token/manager.go`
- **Refresh Token Store**: Redis-backed cache implementing `internal/cache/RefreshTokenStore` interface
- **Database**: PostgreSQL for user accounts with bcrypt password hashing
- **Migrations**: Goose for schema versioning

## Notes
- Access tokens are short-lived JWTs (default: 15 minutes)
- Refresh tokens are long-lived random tokens stored in Redis (default: 7 days)
- Passwords are hashed using bcrypt
- In development mode, migrations run automatically at startup

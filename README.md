# Sipi Backend

Single backend service for an intelligent meeting planning system built as a modular monolith.

## Stack

- Go 1.25
- Gin
- GORM
- PostgreSQL
- Swag / Swagger
- JWT access token + refresh token

## Architecture

- one backend service, no microservices
- layered modules: `handler -> service -> repository`
- PostgreSQL persistence via GORM
- auto-migrate on application startup
- Swagger docs served from `/swagger/index.html`
- calendar import only through `.ics` file upload

## Project Structure

- `cmd/api` - application entrypoint
- `internal/config` - env-based configuration
- `internal/app` - bootstrap, dependency assembly, auto-migrate, shutdown
- `internal/platform` - shared infra (`db`, `jwt`, `httpx`, `logger`)
- `internal/http` - routes and middleware
- `internal/modules` - business modules
- `docs` - generated Swagger docs
- `migrations` - migration-related notes/placeholders
- `scripts/postgres/init` - PostgreSQL init SQL for local Docker demo
- `build` - compiled binaries

## Configuration

Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

Main variables:

- `APP_PORT` - HTTP port, default `8080`
- `POSTGRES_DSN` - PostgreSQL DSN
- `JWT_SECRET` - secret for signing JWT
- `JWT_ACCESS_TTL_MINUTES` - access token TTL
- `JWT_REFRESH_TTL_HOURS` - refresh token TTL

## Local Run Without Docker

1. Start PostgreSQL locally
2. Copy env file
3. Run:

```bash
make tidy
make fmt
make swag
make run
```

Service will be available at:

- API: `http://localhost:8080`
- Swagger UI: `http://localhost:8080/swagger/index.html`
- Health: `http://localhost:8080/health`

## Local Run With Docker Compose

Start full demo environment:

```bash
make docker-up
```

Stop environment:

```bash
make docker-down
```

Services:

- `app` on `http://localhost:8080`
- `postgres` on `localhost:5432`
- `redis` is optional and available only with profile `with-redis`

Run with optional Redis profile:

```bash
docker compose --profile with-redis up --build -d
```

## Swagger

Swagger docs are generated into `docs/`.

Generate or refresh docs:

```bash
make swag
```

Direct command:

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g ./cmd/api/main.go -o ./docs
```

Swagger UI endpoint:

```text
GET /swagger/index.html
```

## Database

Project currently uses **auto-migrate** on startup from [`internal/app/migrator.go`](/home/taa/Рабочий стол/перенос/перенос/sipi/internal/app/migrator.go).

This is the active local/demo migration mechanism:

- app starts
- connects to PostgreSQL
- runs GORM `AutoMigrate(...)`
- creates/updates required tables

Docker PostgreSQL also runs init SQL from:

- [`scripts/postgres/init/001-extensions.sql`](/home/taa/Рабочий стол/перенос/перенос/sipi/scripts/postgres/init/001-extensions.sql)

## Main Make Targets

```bash
make run
make build
make test
make fmt
make swag
make docker-up
make docker-down
```

## Main Endpoint Groups

Public infrastructure:

- `GET /health`
- `GET /swagger/index.html`

Authentication:

- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`
- `POST /auth/logout`
- `POST /auth/logout-all`

Users / profile / availability:

- `GET /users/me`
- `PATCH /users/me`
- `GET /users/me/working-hours`
- `PUT /users/me/working-hours`
- `GET /users/me/unavailability`
- `POST /users/me/unavailability`
- `DELETE /users/me/unavailability/:id`

Events:

- `POST /events`
- `GET /events`
- `GET /events/:id`
- `PATCH /events/:id`
- `DELETE /events/:id`

Calendar import:

- `POST /calendar-import/ics`
- `GET /calendar-import/history`
- `GET /calendar-import/history/:id`

Resources:

- `POST /resources`
- `GET /resources`
- `GET /resources/:id`
- `PATCH /resources/:id`
- `DELETE /resources/:id`
- `GET /resources/:id/availability`
- `POST /resources/:id/bookings`
- `DELETE /resources/:id/bookings/:bookingId`

Meetings:

- `POST /meetings`
- `GET /meetings`
- `GET /meetings/:id`
- `PATCH /meetings/:id`
- `DELETE /meetings/:id`
- `POST /meetings/:id/participants`
- `DELETE /meetings/:id/participants/:userId`
- `POST /meetings/:id/resources`
- `DELETE /meetings/:id/resources/:resourceId`
- `POST /meetings/:id/respond`
- `POST /meetings/:id/request-alternative`

Scheduling:

- `POST /meetings/:id/search-slots`
- `GET /meetings/:id/slots`
- `POST /meetings/:id/select-slot`

Notifications:

- `GET /notifications`
- `POST /notifications/:id/read`
- `POST /notifications/read-all`

## Build

```bash
make build
```

Binary will be created in:

- `./build/sipi-api`

## Tests

```bash
make test
```

## Demo Flow

1. Register user via `/auth/register`
2. Login and get JWT pair
3. Open `/swagger/index.html`
4. Authorize with `Bearer <access_token>`
5. Create profile / working hours / unavailability
6. Create events, resources, meetings
7. Search and select meeting slots
8. Check generated notifications

## Notes

- Refresh tokens are stored in DB only in hashed form
- Passwords are stored only as Argon2id hashes
- Imported calendar events are stored in the shared `events` table with `source_type=imported`
- Recurring meetings and recurring ICS events are implemented as MVP-level support with room for future expansion

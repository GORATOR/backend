# GORATOR Backend

**GORATOR** is a self-hosted platform for tracking errors and exceptions in your applications. 
A Sentry alternative that you deploy on your own server and fully control.

## Why GORATOR?

Production errors are inevitable. The question is how quickly you learn about them and how convenient it is to analyze them.

GORATOR receives error events from your applications, groups them into issues by type and stacktrace, and provides an API for analysis. 
Instead of digging through logs, you get a clear picture: what errors are occurring, how often, in what environment, and in what context.

## How it works

1. You connect a Sentry SDK to your application and specify the DSN of your GORATOR server
2. When an error occurs, the SDK sends an event (envelope) to the `/api/{project_id}/envelope/` endpoint
3. The backend parses the event, extracting exception and stacktrace information
4. Events are grouped into issues by stacktrace hash
5. Through the API and web interface, you analyze errors, filter by project, and track trends

## Tech stack

- **Go 1.22**
- **PostgreSQL 16** — database
- **GORM** — ORM for database operations
- **REST API** with session-based authentication

## Quick start

### Requirements

- Docker and Docker Compose

### Running

```bash
cp .env.example .env
# Fill in the environment variables (see section below)
docker compose -f docker-compose.yml up -d
```

After the containers are up, run migrations:

```bash
docker exec -it <backend-container> ./backend -s
```

The `-s` flag creates all tables and relations. In debug mode (`GORATOR_IS_DEBUG=1`), it also creates test data: 
an organization, a team, and users with admin and viewer roles.

### Running in development mode

```bash
cp .env.example .env
# Fill in the environment variables
docker compose -f docker-compose.yml run -p 5432:5432 db
# Start the application (e.g., F5 in VSCode)
```

## Environment variables

### Database (PostgreSQL container)

| Variable | Description |
|---|---|
| `POSTGRES_DB` | Database name |
| `POSTGRES_USER` | Database user |
| `POSTGRES_PASSWORD` | Database password |

### Database connection (application)

| Variable | Description |
|---|---|
| `GORATOR_DB_HOSTNAME` | PostgreSQL host |
| `GORATOR_DB_PORT` | PostgreSQL port |
| `GORATOR_DB_USERNAME` | Username for database connection |
| `GORATOR_DB_PASSWORD` | Password for database connection |

### Application

| Variable | Description |
|---|---|
| `GORATOR_ALLOWED_ORIGIN` | Allowed origin URL for CORS |
| `GORATOR_SALT` | Salt for password hashing |

### Debug mode

| Variable | Description |
|---|---|
| `GORATOR_IS_DEBUG` | Enable debug mode (`1` — enabled) |
| `GORATOR_DEBUG_USERS_PASSWORD` | Test users password during migration (requires `GORATOR_IS_DEBUG=1`) |
| `GORATOR_SKIP_AUTH_CHECK` | Skip session check, use `GORATOR_AUTH_USER_ID` instead (requires `GORATOR_IS_DEBUG=1`) |
| `GORATOR_AUTH_USER_ID` | User ID for unauthenticated requests (requires `GORATOR_IS_DEBUG=1` and `GORATOR_SKIP_AUTH_CHECK=1`) |

## Testing API without authentication

In debug mode, you can send requests on behalf of any user, bypassing authentication:

```bash
# In .env:
GORATOR_IS_DEBUG=1
GORATOR_SKIP_AUTH_CHECK=1
GORATOR_AUTH_USER_ID=7  # desired user ID

# Example requests (files in the curl/ directory):
curl http://localhost:8080/user -d @curl/create/user.json -X POST -v
curl http://localhost:8080/user -d @curl/update/user.json -X PUT -v
```

## API endpoints

### Authentication
- `POST /login` — sign in, obtain session
- `GET /user/current` — current user

### Event ingestion
- `POST /api/{project_id}/envelope/` — receive errors from SDK (DSN)

### Data
- `GET /envelopes` — list events (pagination, filtering)
- `GET /issues-aggregated` — aggregated issues list
- `GET /issue/{id}/events` — events for a specific issue
- `GET /issue/{id}/events/stats` — issue event statistics
- CRUD for entities: User, Organization, Team, Project, Role

## Related repositories

- [GORATOR Frontend](https://github.com/GORATOR/frontend) — web interface (Vue 3, TypeScript)

# GORATOR Backend

**GORATOR** is a self-hosted platform for tracking errors and exceptions in your applications. 
A Sentry alternative that you deploy on your own server and fully control.

![demo](https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/demo.gif)

## Why GORATOR?

- Self-hosted — full control over your data
- Sentry-compatible SDKs without the operational overhead
- Automatic error grouping by stacktrace
- Simple REST API for analysis
- No need to dig through logs


<p align="center">
  <a href="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/1_issues_filter.png">
    <img src="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/thumb_1_issues_filter.png" />
  </a>
  <a href="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/2_filtered_issues.png">
    <img src="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/thumb_2_filtered_issues.png" />
  </a>
  <a href="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/3_dsn_copy.png">
    <img src="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/thumb_3_dsn_copy.png" />
  </a>
  <a href="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/4_user_profile.png">
    <img src="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/thumb_4_user_profile.png" />
  </a>
  <a href="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/5_rbac_admin.png">
    <img src="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/thumb_5_rbac_admin.png" />
  </a>
  <a href="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/6_team_page.png">
    <img src="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/thumb_6_team_page.png" />
  </a>
  <a href="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/7_issue_details_1.png">
    <img src="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/thumb_7_issue_details_1.png" />
  </a>
  <a href="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/8_issue_details_2.png">
    <img src="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/thumb_8_issue_details_2.png" />
  </a>
  <a href="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/9_issue_details_3.png">
    <img src="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/thumb_9_issue_details_3.png" />
  </a>
  <a href="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/10_issue_details_4.png">
    <img src="https://raw.githubusercontent.com/GORATOR/static_files/refs/heads/main/thumb_10_issue_details_4.png" />
  </a>
</p>

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

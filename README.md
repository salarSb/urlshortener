# URL Shortener — Go, Gin, GORM, Postgres, Docker 🚀

A minimal but solid URL shortener service built with:

- **Go** + **Gin** (HTTP API)
- **GORM** + **PostgreSQL** (persistence)
- **Docker** + **docker-compose** (app + DB)
- 5-character short codes using `[a-zA-Z0-9]`
- Optional expiry + click counting

---

## 🧱 Features

- Shortens any valid URL into a **5-character** short code (e.g. `aB9Zk`)
- Character set: **uppercase + lowercase letters + digits** (`[a-zA-Z0-9]`)
- Optional expiration (in seconds)
- Click count tracking (incremented on each redirect)
- Clean GORM model & migrations via `AutoMigrate`
- Fully dockerized (app + Postgres)

---

## 📁 Project structure

```text
.
├── cmd/
│   └── api/
│       └── main.go          # App entrypoint
├── internal/
│   └── shortener/
│       ├── handler.go       # HTTP handlers (Gin)
│       ├── model.go         # GORM models
│       └── repository.go    # Data access layer
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

---

## ⚙️ Requirements

- **Docker** & **docker-compose** (recommended way to run)
- (Optional) **Go 1.24+** if you want to run it locally without Docker

---

## 🚀 Quick start (Docker)

Clone the repo:

```bash
git clone <YOUR_REPO_URL> urlshortener
cd urlshortener
```

Build and start the stack:

```bash
docker-compose up --build
```

This will start:

- `db`  — Postgres on port **5432**
- `app` — Gin HTTP server on port **8080**

Once everything is up, the API is available at:

```text
http://localhost:8080
```

The `BASE_URL` used in responses is configured to `http://localhost:8080` by default (see `docker-compose.yml`).

---

## 🌍 API

### 1. Create a short URL

**Endpoint**

```http
POST /api/v1/shorten
Content-Type: application/json
```

**Request body**

```json
{
  "url": "https://golang.org",
  "expires_in_seconds": 86400
}
```

- `url` (string, required): a valid URL to shorten  
- `expires_in_seconds` (integer, optional): how long the link is valid (in seconds). If omitted or `<= 0`, the link never expires.

**Example with `curl`:**

```bash
curl -X POST http://localhost:8080/api/v1/shorten   -H "Content-Type: application/json"   -d '{
    "url": "https://golang.org",
    "expires_in_seconds": 86400
  }'
```

**Sample response:**

```json
{
  "short_url": "http://localhost:8080/aB9Zk",
  "short_code": "aB9Zk",
  "original_url": "https://golang.org",
  "expires_at": "2025-11-10T13:30:02.123456Z"
}
```

---

### 2. Redirect via short code

**Endpoint**

```http
GET /:code
```

Where `:code` is the 5-character short code (e.g. `aB9Zk`).

**Example:**

```bash
curl -v http://localhost:8080/aB9Zk
```

You’ll receive a `307 Temporary Redirect` to the original URL.  

If the link is expired, the API returns:

```json
{
  "error": "link expired",
  "originalUrl": "https://golang.org"
}
```

---

## 🧩 Configuration (environment variables)

These env vars are already set in `docker-compose.yml`, but you can override them.

| Variable      | Default                  | Description                                  |
|---------------|--------------------------|----------------------------------------------|
| `DB_HOST`     | `db`                     | DB host (Docker service name)                |
| `DB_PORT`     | `5432`                   | DB port                                      |
| `DB_USER`     | `urlshortener`           | DB username                                  |
| `DB_PASSWORD` | `password`               | DB password                                  |
| `DB_NAME`     | `urlshortener`           | DB name                                      |
| `BASE_URL`    | `http://localhost:8080`  | Base URL used when returning `short_url`     |
| `SERVER_PORT` | `8080`                   | Port Gin listens on                          |

If you run locally (without Docker), set these via your shell or `.env` file.

---

## 🧪 Run locally (without Docker)

If you prefer to run the app directly with Go:

1. Make sure Postgres is running locally and create a DB `urlshortener`:

   ```bash
   createdb urlshortener
   ```

2. Export env vars (example):

   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=youruser
   export DB_PASSWORD=yourpassword
   export DB_NAME=urlshortener
   export BASE_URL=http://localhost:8080
   export SERVER_PORT=8080
   ```

3. Run the app:

   ```bash
   go run ./cmd/api
   ```

Now use the same API calls as in the Docker section.

---

## 📌 Notes

- Database schema is managed automatically by **GORM AutoMigrate** in `main.go` on startup.
- Short codes are generated randomly; we retry a few times if there is a rare unique constraint collision.
- Short code space: `62^5 = 916,132,832` unique codes.

Happy shortening! ✂️🔗

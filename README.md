cat > README.md << 'EOF'
# 🕷️ Secure Web Crawling Feed Aggregation Service



A Go backend service that lets users subscribe to content sources, trigger crawling, and retrieve a personalized aggregated article feed. Built as a learning project focused on backend engineering, Go concurrency, and real-world system design patterns.

---

## Overview

This project implements a backend API where:

- Users register and authenticate with JWT tokens
- Each user subscribes to one or more content source URLs (RSS feeds, Atom feeds, or plain HTML pages)
- Crawls are triggered on demand per-subscription or for all subscriptions at once
- The crawler fetches the source, detects the format (RSS, Atom, or HTML), extracts article data, deduplicates content, and persists it to MongoDB
- Users can retrieve a paginated feed of articles from all their subscribed sources

The server exposes a REST API built with [Gin](https://github.com/gin-gonic/gin), backed by MongoDB, and uses Go goroutines to run crawl jobs asynchronously so the API stays non-blocking.

---

## Why I Built This

I built this while learning Go because I wanted to move beyond basic CRUD applications and explore backend engineering problems that show up in real systems — authentication, background processing, content deduplication, and data aggregation.

Feed aggregation is a good domain for this: it involves HTTP crawling, multiple content formats, async processing, user-scoped data, and some thought around indexing and storage design. It also gave me a concrete reason to learn Go's concurrency model in practice rather than just reading about it.

The goal was not to build something production-ready, but to build something complete enough to understand how these pieces fit together.

---

## Features

- **JWT Authentication** — Signup/login with bcrypt password hashing; tokens carry user identity and role
- **Role-based access** — `ADMIN` users can list all users; regular `USER` accounts are scoped to their own data
- **Source management** — Sources are shared across users; subscribing to an existing URL reuses the same source record
- **Multi-format crawling** — Supports RSS 2.0, Atom, and HTML pages (with site-specific extractors for Hacker News and Lobsters, plus generic `<article>` and heading-link fallbacks)
- **Async crawl dispatch** — Crawl jobs run in background goroutines; the API responds immediately with a "crawl started" acknowledgement
- **Content deduplication** — Checks both URL and SHA-256 content hash before inserting; duplicate articles are skipped
- **Storage cap per source** — A cleanup step keeps the 50 most recent articles per source to bound storage growth
- **Paginated feed API** — Returns articles from subscribed sources sorted by discovery time, with configurable page and limit
- **Indexed collections** — MongoDB indexes on URLs, content hashes, source IDs, and timestamps for efficient lookups
- **Docker support** — `Dockerfile` and `docker-compose.yml` included for running the API and MongoDB together

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.24 |
| HTTP framework | [Gin](https://github.com/gin-gonic/gin) |
| Database | MongoDB 8.0 (via `mongo-driver`) |
| Authentication | JWT (`dgrijalva/jwt-go`) + bcrypt |
| HTML parsing | [goquery](https://github.com/PuerkitoBio/goquery) |
| XML parsing | Go standard library (`encoding/xml`) |
| Validation | `go-playground/validator` |
| Config | `godotenv` |
| Containerisation | Docker + Docker Compose |

---

## Architecture / Project Structure

```
.
├── main.go                    # Entry point: server setup, route registration, index creation
├── database/
│   ├── databaseConnection.go  # MongoDB client initialisation
│   └── indexes.go             # Index creation for all collections at startup
├── models/
│   ├── userModels.go          # User struct (email, password hash, tokens, role)
│   ├── sourceModels.go        # Source struct (URL, status, crawl stats, change-detection fields)
│   ├── subscriptionModels.go  # Subscription struct (user ↔ source relationship)
│   └── articleModels.go       # Article struct (title, URL, content hash, summary, author)
├── routes/
│   ├── authRoutes.go          # Public routes: POST /users/signup, POST /users/login
│   ├── userRoutes.go          # Protected routes: GET /users, GET /users/:user_id
│   └── subscriptionRouter.go  # Protected routes: subscriptions, crawl, feed
├── controllers/
│   ├── userController.go      # Signup, Login, GetUser, GetUsers handlers
│   ├── subscriptionController.go # Add, Remove, List subscription handlers
│   ├── crawlerController.go   # CrawlSubscription, CrawlAllSubscriptions handlers
│   └── feedsController.go     # GetFeed handler (paginated)
├── services/
│   ├── subscriptionService.go # Business logic: add/remove/list subscriptions; source upsert
│   ├── crawlerService.go      # Crawl orchestration: fetch → extract → deduplicate → save → cleanup
│   ├── extractorService.go    # Content extraction: RSS, Atom, HTML parsers
│   └── feedService.go         # Feed assembly: fetch articles from subscribed sources
├── middleware/
│   └── authMiddleware.go      # JWT validation; sets user claims on Gin context
├── helpers/
│   ├── tokenHelper.go         # JWT generation, validation, token refresh in DB
│   ├── hashHelper.go          # SHA-256 content hash generation
│   └── authHelper.go          # Role checks (CheckUserType, MatchUserTypeToUid)
├── Dockerfile
├── docker-compose.yml
└── .env
```

**Collections in MongoDB:**

| Collection | Purpose |
|---|---|
| `user` | User accounts |
| `sources` | Crawlable URLs (shared, one record per unique URL) |
| `subscriptions` | User ↔ source mappings |
| `articles` | Crawled article data |

---

## How It Works

### Authentication

On signup, the password is hashed with bcrypt (cost factor 14) and a JWT access token (24h expiry) plus a refresh token (168h) are generated and stored in the user document. On login, the existing tokens are regenerated and updated in the database. All protected routes pass through the `Authenticate` middleware, which reads the token from the `token` request header, validates the signature and expiry, and attaches the user's claims to the Gin context.

### Subscription & Source Management

When a user subscribes to a URL:
1. The URL is validated and normalised (trailing slash stripped).
2. The `sources` collection is checked for an existing record with that URL. If none exists, a new source document is inserted (using the hostname as the default display name).
3. A `subscriptions` record linking the user to the source is created. A compound unique index on `(user_id, source_id)` prevents duplicate subscriptions.

Sources are shared — if two users subscribe to the same URL, they reference the same source document and share crawled articles.

### Crawling Pipeline

Crawl jobs are dispatched as goroutines and run in the background. The request handler returns immediately.

Inside `CrawlSource`:

1. Load the source record from MongoDB.
2. Update `last_attempt_at`.
3. Fetch the URL using `net/http` with a 30-second timeout and a browser-like `User-Agent`.
4. Detect content type by inspecting the response body for `<rss`, `<channel>`, `<feed`, or Atom namespace markers.
5. Attempt parsers in order: RSS → Atom → HTML fallback.
6. For each extracted article, check for duplicate URL (per source) and duplicate content hash (globally). Skip duplicates.
7. Insert new articles. Track count of saved articles.
8. Run a cleanup step: if the source now has more than 50 articles, delete the oldest ones by `discovered_at`.
9. Update the source record with `status`, `last_crawled_at`, and crawl counters.

### Content Extraction

| Parser | Format | Library |
|---|---|---|
| `parseRSSFeed` | RSS 2.0 (`<rss>`) | `encoding/xml` |
| `parseAtomFeed` | Atom (`<feed>`) | `encoding/xml` |
| `parseHTMLPage` | Generic HTML | `goquery` |

The HTML parser tries strategies in order: Hacker News-specific selectors, Lobsters-specific selectors, `<article>` tags, common class names (`div.post`, `div.entry`, etc.), and finally `h1/h2/h3 > a` link extraction.

Summaries are stripped of HTML tags and capped at 500 characters. Dates are tried against a list of common formats (RFC1123, RFC3339, RFC822, and several common variants).

A SHA-256 hash of `title + summary` is stored as `content_hash` on each article, used for cross-source deduplication.

### Feed Delivery

`GET /api/feed` collects the user's subscribed source IDs, queries the `articles` collection filtered by those IDs, sorts by `discovered_at` descending, and applies skip/limit for pagination. Each result is a combined article + source object so the client knows which source the article came from.

---

## API Endpoints Summary

### Authentication (public)

| Method | Path | Description |
|---|---|---|
| `POST` | `/users/signup` | Create a new user account |
| `POST` | `/users/login` | Login and receive a JWT token |

**Signup request body:**
```json
{
  "first_name": "Jane",
  "last_name": "Doe",
  "email": "jane@example.com",
  "password": "secret123",
  "phone": "9876543210",
  "user_type": "USER"
}
```

`user_type` must be `USER` or `ADMIN`.

**Login request body:**
```json
{
  "email": "jane@example.com",
  "password": "secret123"
}
```

Login returns the full user object including the `token` field. Use that value as the `token` header on all subsequent requests.

---

### Users (requires `token` header)

| Method | Path | Access | Description |
|---|---|---|---|
| `GET` | `/users` | ADMIN only | List all users (paginated) |
| `GET` | `/users/:user_id` | Own user or ADMIN | Get a specific user by ID |

---

### Subscriptions (requires `token` header)

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/subscriptions` | Subscribe to a new source URL |
| `GET` | `/api/subscriptions` | List all subscriptions (with source details) |
| `DELETE` | `/api/subscriptions/:id` | Remove a subscription |

**Subscribe request body:**
```json
{
  "url": "https://feeds.example.com/rss"
}
```

---

### Crawl (requires `token` header)

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/crawl/:id` | Trigger a crawl for a specific subscription |
| `POST` | `/api/crawl/all` | Trigger a crawl for all subscriptions |

Both endpoints return immediately. The crawl runs in a background goroutine.

---

### Feed (requires `token` header)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/feed` | Get paginated articles from subscribed sources |

**Query parameters:** `page` (default: 1), `limit` (default: 20, max: 100)

**Response:**
```json
{
  "page": 1,
  "limit": 20,
  "total": 87,
  "total_pages": 5,
  "articles": [
    {
      "article": { "id": "...", "title": "...", "url": "...", "summary": "...", "published_at": "..." },
      "source":  { "id": "...", "name": "...", "url": "...", "status": "active" }
    }
  ]
}
```

---

## Getting Started

### Prerequisites

- Go 1.21+
- MongoDB (local instance or the Docker Compose setup below)

### Option 1: Run with Docker Compose (recommended)

```bash
git clone <repo-url>
cd Go-lang-JWT

# Copy and edit the environment file
cp .env.example .env
# Set a strong SECRET_KEY in .env

docker compose up --build
```

The API will be available at `http://localhost:8080`. MongoDB data is persisted in a named Docker volume.

### Option 2: Run locally

```bash
git clone <repo-url>
cd Go-lang-JWT

cp .env.example .env
# Edit .env:
#   MONGODB_URI=mongodb://localhost:27017/cluster0
#   SECRET_KEY=your-secret-key
#   PORT=8080

go run main.go
```

### Environment Variables

| Variable | Description | Example |
|---|---|---|
| `PORT` | Server port | `8080` |
| `MONGODB_URI` | MongoDB connection string | `mongodb://localhost:27017/cluster0` |
| `SECRET_KEY` | JWT signing secret | any strong random string |

---

## Example Workflow

```bash
BASE_URL="http://localhost:8080"

# 1. Create an account
curl -s -X POST "$BASE_URL/users/signup" \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Jane","last_name":"Doe","email":"jane@example.com","password":"secret123","phone":"9876543210","user_type":"USER"}'

# 2. Log in and extract the token
TOKEN=$(curl -s -X POST "$BASE_URL/users/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"jane@example.com","password":"secret123"}' | jq -r '.token')

# 3. Subscribe to an RSS feed
curl -s -X POST "$BASE_URL/api/subscriptions" \
  -H "token: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://hnrss.org/frontpage"}'

# 4. Trigger a crawl for all subscriptions
curl -s -X POST "$BASE_URL/api/crawl/all" \
  -H "token: $TOKEN"

# 5. Fetch the feed (wait a few seconds for the crawl to complete)
curl -s "$BASE_URL/api/feed?page=1&limit=10" \
  -H "token: $TOKEN" | jq '.articles[].article.title'
```

---

## Current Status

This project is **functionally complete** for its intended scope and runs correctly in a local environment. It is not deployed and has not been load tested or hardened for production use.

What works:
- Full signup/login/token authentication flow
- Subscription management with source deduplication
- RSS, Atom, and HTML crawling with content deduplication
- Background crawl dispatch via goroutines
- Paginated feed retrieval
- MongoDB index setup on startup
- Docker Compose for local development

---

## Limitations

- **No automatic / scheduled crawling** — Crawls are user-triggered. There is no background scheduler or cron job that refreshes sources periodically.
- **No rate limiting** — There is no per-user or per-IP rate limiting on any endpoint, including crawl triggers.
- **JWT stored in database** — Tokens are persisted to the user document. There is no token blacklist or revocation mechanism; logout is not implemented.
- **Token header, not `Authorization: Bearer`** — The middleware reads from a `token` header, which is not the standard HTTP auth header format.
- **Sequential crawl for `/api/crawl/all`** — `CrawlUserSources` iterates over subscriptions and crawls them one by one, which can be slow with many subscriptions. There is no worker pool or concurrency limit on the per-source crawls.
- **HTML extraction is best-effort** — The HTML parser works for Hacker News and Lobsters specifically, and has generic fallbacks, but it will not reliably extract articles from all sites.
- **No input sanitisation beyond validation** — Article content is stored as-is from the source with no further sanitisation.
- **`log.Panic` in some error paths** — Some error paths (e.g., database failures during signup) use `log.Panic`, which will crash the process.

---

## Future Improvements

- [ ] Add a background scheduler (e.g., using `time.Ticker` or a job queue) to refresh sources automatically on a configurable interval
- [ ] Introduce a worker pool with bounded concurrency for `CrawlUserSources` to parallelize multi-source crawls safely
- [ ] Replace the custom `token` header with standard `Authorization: Bearer <token>` format
- [ ] Implement token revocation or a stateless refresh-token flow
- [ ] Add per-user and global rate limiting (e.g., via a Gin middleware)
- [ ] Expose source-level metadata in the feed response (crawl status, last crawled time) so clients can show freshness information
- [ ] Add a `GET /api/sources/:id/articles` endpoint for source-specific browsing
- [ ] Write unit tests for the service layer, particularly the extractor and deduplication logic
- [ ] Replace `log.Panic` with structured error handling that returns appropriate HTTP responses instead of crashing
- [ ] Add structured logging (e.g., `slog` or `zap`) for better observability in a multi-goroutine environment

---

## License

This project was built for personal learning and is not actively maintained. Feel free to use it as a reference.


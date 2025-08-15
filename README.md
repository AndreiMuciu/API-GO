# API-GO

An opinionated Go REST API using Gorilla Mux and MongoDB, with a clean architecture (handlers → services → repositories) and practical helpers for auth, validation, responses, and generic CRUD routing.

## Stack

- Go + Gorilla Mux (routing)
- MongoDB Go Driver (persistence)
- JWT (github.com/golang-jwt/jwt/v5)
- Bcrypt (golang.org/x/crypto/bcrypt)

## Highlights

- Modular auth with JWT stored as HttpOnly cookie (SameSite=Lax; Secure flag via env).
- Repository pattern for persistence; services encapsulate business logic; handlers focus on HTTP.
- Generic CRUD router helper to wire standard endpoints fast.
- Reusable query parser: filtering, sorting, pagination for list endpoints (used for Books).
- Parallelism for better latency on independent DB operations.

## Project structure (key parts)

- `cmd/server/main.go` – wiring of config, DB, repositories, services, routers, and HTTP server.
- `internal/config` – environment-based configuration.
- `internal/database` – Mongo connection and indexes.
- `internal/models` – domain models and DTOs.
- `internal/repository` – interfaces and Mongo implementations.
- `internal/services` – business logic (e.g., authentication).
- `internal/handlers` – HTTP controllers.
- `internal/router` – per-domain routers and a generic CRUD mount.
- `internal/utils` – JWT, password hashing, responses, validation, query parsing.

## Configuration

Set these environment variables (examples in parentheses):

- `PORT` (e.g., `8080`)
- `MONGO_URI` (e.g., `mongodb+srv://user:<db_password>@cluster/...`)
- `DB_PASSWORD` – substituted into `MONGO_URI` in place of `<db_password>`
- `JWT_SECRET` – secret for signing JWTs
- `JWT_TTL_MINUTES` – access token TTL in minutes (default 60)
- `COOKIE_NAME` – auth cookie name (default `access_token`)
- `COOKIE_SECURE` – `true|false` to mark cookie Secure (default false for local)

## Run (dev)

1. Ensure Mongo is reachable; set env vars above.
2. Start server:

```powershell
go run cmd/server/main.go
```

Server listens on `http://localhost:<PORT>` and mounts API under `/api-go/v1`.

## Routes

Base prefix for all endpoints: `/api-go/v1`.

### Auth

- POST `/auth/signup` – Register user; sets JWT cookie on success.
- POST `/auth/login` – Login; sets JWT cookie on success.
- POST `/auth/logout` – Logout; clears JWT cookie.

Request DTOs:

- Signup: `{ name, email, password, passwordConfirm, phone? }`
- Login: `{ email, password }`

Cookie details:

- HttpOnly, SameSite=Lax, `Secure` from `COOKIE_SECURE`, expiry from `JWT_TTL_MINUTES`.

### Users

- GET `/users` – List users.
- GET `/users/{id}` – Get one user.
- PUT `/users/{id}` – Update selected fields (name, email, phone) with validation and uniqueness checks.
- DELETE `/users/{id}` – Delete user.

Notes:

- Creating users via `POST /users` is disabled. Use `/auth/signup`.
- Passwords are not returned in responses.

### Books (CRUD + filtering/sorting/pagination)

- GET `/books` – List books with filters; returns `{ items, page, limit, total }`.
- GET `/books/{id}` – Get one book.
- POST `/books` – Create a book.
- PUT `/books/{id}` – Update selected fields.
- DELETE `/books/{id}` – Delete a book.

Book model: `{ id, title, author, yearPublished, genre }`.

List query parameters (allowlisted fields: title, author, genre, yearPublished):

- Equality: `?author=Asimov&genre=Sci-Fi`
- Contains (case-insensitive): `?title_like=foundation`
- Numeric ranges: `?yearPublished_min=1950&yearPublished_max=1970`
- Global search across string fields: `?q=scifi`
- Sorting: `?sort=yearPublished,-title` (leading `-` = desc)
- Pagination: `?page=2&limit=10` (defaults: sort by `title`, limit `20`, max `100`)

## How it works

### Architecture flow

1. Router maps HTTP routes to handlers (controllers).
2. Handlers validate/parse input and call services/repositories.
3. Services implement business logic (e.g., auth signup/login), using repositories for data.
4. Repositories hide persistence details (MongoDB in this case).
5. Utils provide cross-cutting helpers (JWT, password, responses, validation, query parsing).

### Authentication

- Signup flow: validate -> parallel checks for email/phone uniqueness -> hash password -> create user -> generate JWT -> set cookie.
- Login flow: fetch by email -> verify password -> generate JWT -> set cookie.

### Users

- Update flow: validate fields -> check email/phone uniqueness in parallel if provided -> update only provided fields.

### Books

- Listing flow: parse filters/sort/pagination -> repository runs Find and CountDocuments in parallel -> return items + meta.

## Parallelism and performance

- Auth signup: EmailExists and PhoneExists checks run concurrently.
- Users update: EmailExists and PhoneExists checks run concurrently when both fields are present.
- Books list: fetching items and counting total run in parallel.
- Unique indexes on `users.email` (unique) and `users.phone` (unique sparse) protect against race conditions at DB level.

## Responses and errors

All responses use a consistent JSON shape via `internal/utils/response.go`:

- Success: `{ success: true, message, data }`
- Error: `{ success: false, error }` with appropriate HTTP status codes (400, 401, 403, 404, 405, 409, 422, 429, 500, 503, etc.).

## Extending with a new CRUD resource

1. Create your model in `internal/models`.
2. Create a repository interface + Mongo implementation in `internal/repository`.
3. Create a handler implementing the `CRUDHandlers` interface in `internal/handlers`.
4. Create a router that calls `router.MountCRUD(r, "/<resource>", handler)`.
5. Mount the router in `cmd/server/main.go` under `/api-go/v1/<resource>`.
6. (Optional) For list endpoints, use `utils.ParseListQuery` and implement `ListWithQuery` in the repository.

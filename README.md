# ✅ Todo API

A robust, production-ready REST API for managing todo lists with JWT authentication. Built with Go and PostgreSQL, emphasizing clean architecture, security, and developer experience.

**Live API:** `https://todo-api-production-74d1.up.railway.app`

**📖 API Documentation:** [Swagger UI](https://todo-api-production-74d1.up.railway.app/swagger/index.html) — interactive docs, try endpoints directly from your browser

---

## 🚀 Tech Stack

- **Go** — backend language
- **PostgreSQL** — database
- **GORM** — ORM for database operations
- **Architecture** — Layered Architecture (Handler, Middleware, Repository Pattern)
- **JWT (HS256)** — authentication with bcrypt password hashing
- **Validation** — `go-playground/validator/v10`
- **Configuration** — `spf13/viper` (Environment Variables)
- **Logging** — `rs/zerolog` (JSON Structured Logging)
- **Migrations** — `golang-migrate`
- **Swagger/OpenAPI** — auto-generated API documentation
- **Deployment** — Docker & Docker Compose, deployed on Railway

---

## ✨ Features & Engineering Highlights

- **Clean Architecture** — Separation of concerns using the Repository Pattern for maintainability and testability.
- **Robust Authentication** — Secure JWT implementation with `bcrypt` password hashing.
- **Data Isolation** — User-owned resources ensuring users can only access and modify their own todos.
- **Batch Create** — Add multiple todos in one request with per-item validation.
- **Pagination** — `page` and `limit` query parameters with defaults.
- **Production-Ready Logging** — Structured JSON logging (`zerolog`) for easier monitoring and debugging.
- **Graceful Error Handling** — Custom panic recovery middleware catches unexpected errors gracefully.
- **Containerized** — Multi-stage Docker build for a minimal, secure, and optimized container footprint.

---

## 🛣️ Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/register` | ❌ | Register new account |
| `POST` | `/login` | ❌ | Login + get JWT token |
| `POST` | `/tambah-todo` | ✅ | Add a new todo |
| `POST` | `/tambah-todo-batch` | ✅ | Add multiple todos at once |
| `GET` | `/todos?page=1&limit=10` | ✅ | Get all todos (paginated) |
| `PUT` | `/update-todo?id=X` | ✅ | Update a todo |
| `DELETE` | `/hapus-todo?id=X` | ✅ | Delete a todo (soft delete) |

> For detailed request/response schemas, see the [Swagger documentation](https://todo-api-production-74d1.up.railway.app/swagger/index.html).

---

## 💻 Request & Response Examples

### Register
```http
POST /register
Content-Type: application/json

{
    "username": "jason",
    "password": "rahasia123"
}
```

Response:
```json
{"pesan": "Berhasil menambahkan username ke database"}
```

### Login
```http
POST /login
Content-Type: application/json

{
    "username": "jason",
    "password": "rahasia123"
}
```

Response:
```json
{
    "pesan": "Berhasil login!",
    "token": "eyJhbGci..."
}
```

### Add Todo
```http
POST /tambah-todo
Authorization: Bearer eyJhbGci...
Content-Type: application/json

{
    "judul": "Belajar Go",
    "prioritas": "tinggi"
}
```

Response:
```json
{
    "pesan": "Todo berhasil ditambahkan",
    "judul": "Belajar Go",
    "prioritas": "tinggi"
}
```

### Add Batch
```http
POST /tambah-todo-batch
Authorization: Bearer eyJhbGci...
Content-Type: application/json

[
    {"judul": "Belajar Go", "prioritas": "tinggi"},
    {"judul": "Olahraga", "prioritas": "sedang"},
    {"judul": "", "prioritas": "tinggi"}
]
```

Response:
```json
[
    {"judul": "Belajar Go", "prioritas": "tinggi", "status": "berhasil"},
    {"judul": "Olahraga", "prioritas": "sedang", "status": "berhasil"},
    {"judul": "", "error": "judul harus terisi"}
]
```

### Get All Todos
```http
GET /todos?page=1&limit=10
Authorization: Bearer eyJhbGci...
```

Response:
```json
{
    "page": 1,
    "limit": 10,
    "total": 2,
    "data": [
        {"id": 1, "judul": "Belajar Go", "prioritas": "tinggi"},
        {"id": 2, "judul": "Olahraga", "prioritas": "sedang"}
    ]
}
```

---

## 🛡️ Validation Rules

- `judul` — required
- `prioritas` — required, one of: `tinggi`, `sedang`, `rendah`

---

## 🛠️ Local Development

### Option 1: Using Docker (Recommended)
The easiest way to run this project locally is using Docker. It will automatically spin up the API, PostgreSQL database, and run the database migrations.

**Prerequisites:** Docker & Docker Compose

```bash
git clone https://github.com/Tarquished/todo-api.git
cd todo-api
docker-compose up --build
```

### Option 2: Native Execution
**Prerequisites:** Go 1.22+, PostgreSQL

```bash
git clone https://github.com/Tarquished/todo-api.git
cd todo-api
go mod tidy
swag init
go run main.go
```

Server runs at `http://localhost:8080`. Swagger UI at `http://localhost:8080/swagger/index.html`.

---

## ⚙️ Environment Variables

If running natively without Docker, create a `.env` file in the root directory:

| Variable | Example Value | Description |
|----------|-------------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL Host |
| `DB_USER` | `postgres` | Database User |
| `DB_PASSWORD` | `secret123` | Database Password |
| `DB_NAME` | `todo_db` | Database Name |
| `DB_PORT` | `5432` | Database Port |
| `PORT` | `8080` | Server Port |
| `JWT_SECRET` | `your_super_secret_key`| Secret key for JWT signing |

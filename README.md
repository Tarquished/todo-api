# ✅ Todo API

A REST API for managing todo lists with JWT authentication, built with Go and PostgreSQL.

**Live API:** `https://todo-api-production-74d1.up.railway.app`

**📖 API Documentation:** [Swagger UI](https://todo-api-production-74d1.up.railway.app/swagger/index.html) — interactive docs, try endpoints directly from your browser

---

## Tech Stack

- **Go** — backend language
- **PostgreSQL** — database
- **GORM** — ORM for database operations
- **JWT (HS256)** — authentication with bcrypt password hashing
- **Swagger/OpenAPI** — auto-generated API documentation
- **Railway** — cloud deployment

---

## Endpoints

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

## Features

- **JWT Authentication** — register, login, protected endpoints
- **User Ownership** — each user's todos are isolated (filtered by user_id)
- **Batch Create** — add multiple todos in one request with per-item validation
- **Pagination** — `page` and `limit` query parameters with defaults
- **Input Validation** — prioritas must be `tinggi`, `sedang`, or `rendah`
- **Panic Recovery** — middleware catches unexpected errors gracefully
- **Swagger Docs** — interactive API documentation at `/swagger/index.html`

---

## Request & Response Examples

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
    "judul": "Belajar Go",
    "pesan": "Todo berhasil ditambahkan",
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

## Validation Rules

- `judul` — required
- `prioritas` — required, one of: `tinggi`, `sedang`, `rendah`

---

## Local Development

**Prerequisites:** Go 1.22+, PostgreSQL

```bash
git clone https://github.com/Tarquished/todo-api.git
cd todo-api
go mod tidy
swag init
go run main.go
```

Server runs at `http://localhost:8080`. Swagger UI at `http://localhost:8080/swagger/index.html`.

**Environment Variables:**

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `PORT` | Server port (default: 8080) |
| `JWT_SECRET` | Secret key for JWT signing |

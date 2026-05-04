# ✅ Todo API

A REST API for managing todo lists, built with Go and PostgreSQL.

**Live API:** `https://todo-api-production-74d1.up.railway.app`

---

## Tech Stack

- **Go** — backend language
- **PostgreSQL** — database
- **GORM** — ORM for database operations
- **Railway** — cloud deployment

---

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/tambah-todo` | Add a new todo |
| `POST` | `/tambah-todo-batch` | Add multiple todos at once |
| `GET` | `/todos` | Get all todos |
| `PUT` | `/update-todo?id=X` | Update a todo |
| `DELETE` | `/hapus-todo?id=X` | Delete a todo (soft delete) |

---

## Request & Response Examples

### Add Todo
```http
POST /tambah-todo
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
GET /todos
```

Response:
```json
[
    {"id": 1, "judul": "Belajar Go", "prioritas": "tinggi"},
    {"id": 2, "judul": "Olahraga", "prioritas": "sedang"}
]
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
go run main.go
```

Server runs at `http://localhost:8080`.

**Environment Variables:**

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `PORT` | Server port (default: 8080) |
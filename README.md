# Todo API

REST API untuk manage todo list menggunakan Go + PostgreSQL + GORM.

## Endpoint

| Method | URL | Fungsi |
|--------|-----|--------|
| POST | `/tambah-todo` | Tambah todo baru |
| POST | `/tambah-todo-batch` | Tambah banyak todo sekaligus |
| GET | `/todos` | Lihat semua todo |
| PUT | `/update-todo?id=X` | Update todo |
| DELETE | `/hapus-todo?id=X` | Hapus todo (soft delete) |

## Tech Stack

- Go 1.22
- PostgreSQL
- GORM (ORM)
- net/http (web server)

## Local Development

```bash
go mod download
go run main.go
```

Server jalan di `http://localhost:8080`.

## Environment Variables

- `DATABASE_URL` - PostgreSQL connection string
- `PORT` - Server port (default: 8080)

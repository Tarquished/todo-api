# ==========================================
# STAGE 1: BUILDER
# ==========================================
# PENTING: Perhatikan ada "AS builder" di ujung baris ini
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy dependency dan download
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary (CGO_ENABLED=0 penting agar binary bisa jalan mandiri di Alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# ==========================================
# STAGE 2: FINAL (Minimalist)
# ==========================================
FROM alpine:latest

WORKDIR /app

# KITA CUMA COPY FILE BINARY DARI STAGE BUILDER, TINGGALKAN COMPILERNYA
COPY --from=builder /app/server .

# Expose port (sebagai dokumentasi)
EXPOSE 8080

# Jalankan server
CMD ["./server"]
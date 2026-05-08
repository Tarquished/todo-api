# Step 1: Pake Go sebagai base
FROM golang:1.25-alpine

# Step 2: Bikin folder kerja di dalam container
WORKDIR /app

# Step 3: Copy dependency files dulu (biar cache efisien)
COPY go.mod go.sum ./
RUN go mod download

# Step 4: Copy semua kode
COPY . .

# Step 5: Build jadi binary
RUN go build -o server .

# Step 6: Jalanin binary-nya
CMD ["./server"]
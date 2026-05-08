CREATE TABLE IF NOT EXISTS todos (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER,
    judul VARCHAR(255),
    prioritas VARCHAR(255)
);
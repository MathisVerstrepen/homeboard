CREATE TABLE IF NOT EXISTS background (
    id SERIAL PRIMARY KEY,
    filename TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    selected BOOLEAN DEFAULT FALSE
);

INSERT INTO background (filename) VALUES ('background1.jpg');
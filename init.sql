CREATE TABLE IF NOT EXISTS background (
    id SERIAL PRIMARY KEY,
    filename TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    selected BOOLEAN DEFAULT FALSE
);

INSERT INTO background (filename, selected) VALUES ('default.webp', 't');

CREATE TABLE IF NOT EXISTS home_layout (
    position TEXT PRIMARY KEY,
    module_name TEXT NOT NULL
);
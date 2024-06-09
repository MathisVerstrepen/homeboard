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

CREATE TABLE IF NOT EXISTS module_variable (
    position TEXT NOT NULL,
    variable_name TEXT NOT NULL,
    value TEXT NOT NULL,

    CONSTRAINT module_variable_pk PRIMARY KEY (position, variable_name)
);
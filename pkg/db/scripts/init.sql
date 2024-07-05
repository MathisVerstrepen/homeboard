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

-- LinkHub --

CREATE TABLE IF NOT EXISTS linkhub_link (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    icon TEXT,
    is_nsfw BOOLEAN DEFAULT FALSE,
    is_starred BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS linkhub_image (
    linkhub_id INTEGER NOT NULL,
    ext TEXT NOT NULL,
    image_id TEXT NOT NULL PRIMARY KEY,
    is_nsfw BOOLEAN DEFAULT FALSE,

    CONSTRAINT linkhub_image_fk FOREIGN KEY (linkhub_id) REFERENCES linkhub_link(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS linkhub_tag (
    id SERIAL PRIMARY KEY,
    tag TEXT NOT NULL UNIQUE,
    is_nsfw BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS linkhub_link_tag (
    linkhub_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,

    CONSTRAINT linkhub_link_tag_fk FOREIGN KEY (linkhub_id) REFERENCES linkhub_link(id) ON DELETE CASCADE,
    CONSTRAINT linkhub_tag_fk FOREIGN KEY (tag_id) REFERENCES linkhub_tag(id) ON DELETE CASCADE,
    CONSTRAINT linkhub_link_tag_pk PRIMARY KEY (linkhub_id, tag_id)
);

CREATE TABLE IF NOT EXISTS linkhub_image_tag (
    image_id TEXT NOT NULL,
    tag_id INTEGER NOT NULL,

    CONSTRAINT linkhub_image_tag_fk FOREIGN KEY (image_id) REFERENCES linkhub_image(image_id) ON DELETE CASCADE,
    CONSTRAINT linkhub_tag_fk FOREIGN KEY (tag_id) REFERENCES linkhub_tag(id) ON DELETE CASCADE,
    CONSTRAINT linkhub_image_tag_pk PRIMARY KEY (image_id, tag_id)
);
CREATE TABLE IF NOT EXISTS users(
    id SERIAL UNIQUE,
    name TEXT,
    city TEXT,
    state TEXT,
    country TEXT,
    CONSTRAINT pkey_users PRIMARY KEY(id)
);
CREATE TABLE IF NOT EXISTS users(
    id SERIAL,
    name TEXT,
    city TEXT,
    state TEXT,
    country TEXT,
    CONSTRAINT pkey_users PRIMARY KEY(id)
);
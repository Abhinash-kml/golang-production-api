CREATE TABLE IF NOT EXISTS posts(
    id SERIAL UNIQUE,
    title TEXT,
    body TEXT,
    likes INT,
    author_id INT,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT pkey_posts PRIMARY KEY(id)
);
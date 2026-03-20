CREATE TABLE IF NOT EXISTS posts(
    id SERIAL UNIQUE,
    title TEXT,
    body TEXT,
    likes INT,
    authorid INT,
    created_at TIMESTAMP,
    CONSTRAINT pkey_posts PRIMARY KEY(id)
);
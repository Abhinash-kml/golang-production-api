CREATE TABLE IF NOT EXISTS comments(
    id SERIAL UNIQUE,
    author_id SERIAL, 
    post_id SERIAL, 
    body TEXT, 
    likes INT,
    CONSTRAINT pkey_comments PRIMARY KEY(id)
);
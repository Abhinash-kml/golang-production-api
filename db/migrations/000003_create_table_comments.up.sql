CREATE TABLE IF NOT EXISTS comments(
    id SERIAL UNIQUE,
    authorid SERIAL, 
    postid SERIAL, 
    body TEXT, 
    likes INT,
    CONSTRAINT pkey_comments PRIMARY KEY(id)
);
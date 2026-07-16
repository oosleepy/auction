-- +goose Up

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    password_hash VARCHAR(72) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    totalbid INTEGER NOT NULL DEFAULT 0
);

-- +goose Down

DROP TABLE users;



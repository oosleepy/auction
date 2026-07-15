-- +goose Up

CREATE TABLE bid_history (
    id SERIAL PRIMARY KEY,
    bid_name VARCHAR(50) NOT NULL,
    bidamt VARCHAR(50) NOT NULL,
    winner VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down

DROP TABLE bid_history;

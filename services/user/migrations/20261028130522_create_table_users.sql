-- +goose Up
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    phone_number  VARCHAR(50) NOT NULL UNIQUE,
    full_name  VARCHAR(50) NOT NULL,
    address VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE users;

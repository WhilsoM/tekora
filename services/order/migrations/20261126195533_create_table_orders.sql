-- +goose Up
CREATE TABLE orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL CHECK (user_id > 0),
    total_amount BIGINT NOT NULL CHECK(total_amount >= 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'paid', 'cancelled', 'completed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE order_items (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_id BIGINT REFERENCES orders(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL CHECK (product_id >= 0),
    product_name TEXT NOT NULL,
    unit_price BIGINT NOT NULL CHECK(unit_price >= 0),
    quantity INT NOT NULL CHECK(quantity > 0)
);

-- +goose Down
DROP TABLE order_items;
DROP TABLE orders;

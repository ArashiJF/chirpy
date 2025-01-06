-- +goose Up
CREATE TABLE chirps(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    body TEXT UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,

    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE chirps;
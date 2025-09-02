-- +goose up
CREATE TABLE users(
    id UUID  PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    name TEXT NOT NULL
);

-- +goose down
DROP TABLE users;
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS teams (
    team_name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    user_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username  TEXT NOT NULL,
    team_name TEXT NOT NULL REFERENCES teams(team_name),
    is_active BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_team ON users(team_name);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active) WHERE is_active = true;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
DROP INDEX IF EXISTS idx_users_team;
DROP INDEX IF EXISTS idx_users_active;
-- +goose StatementEnd
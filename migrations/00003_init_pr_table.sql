-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id   TEXT PRIMARY KEY,
    pull_request_name TEXT NOT NULL,
    author_id         UUID NOT NULL REFERENCES users(user_id),
    status            TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')) DEFAULT 'OPEN',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    merged_at         TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_pull_requests_author ON pull_requests(author_id);

CREATE TABLE IF NOT EXISTS pr_reviewers (
    pull_request_id TEXT NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    reviewer_id     UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    order_index     SMALLINT NOT NULL CHECK (order_index IN (1, 2)),
    assigned_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (pull_request_id, reviewer_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_pr_order_unique ON pr_reviewers(pull_request_id, order_index);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_user ON pr_reviewers(reviewer_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;

DROP INDEX IF EXISTS idx_pr_reviewers_user;
DROP INDEX IF EXISTS idx_pull_requests_author;
DROP INDEX IF EXISTS idx_pr_order_unique;
-- +goose StatementEnd
-- +goose Up
-- +goose StatementBegin
CREATE TABLE accounts (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  username TEXT NOT NULL,
  address TEXT NOT NULL,
  seed TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX idx_accounts_user_id
ON accounts (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

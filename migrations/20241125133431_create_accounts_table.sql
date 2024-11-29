-- +goose Up
-- +goose StatementBegin
CREATE TABLE accounts (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  address TEXT NOT NULL,
  seed TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_accounts_user_id
ON accounts (user_id);

CREATE UNIQUE INDEX idx_accounts_user_id_address
on accounts (user_id, address);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

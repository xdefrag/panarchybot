-- +goose Up
-- +goose StatementBegin
CREATE TABLE balances (
  user_id bigint NOT NULL,
  asset text NOT NULL,
  pending double precision NOT NULL,
  sent double precision NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

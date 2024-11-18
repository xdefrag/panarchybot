-- name: GetState :one
SELECT * FROM states
WHERE user_id = @user_id
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateState :exec
INSERT INTO states (user_id, state, data, meta, created_at)
  VALUES (@user_id, @state, @data, @meta, now());

-- name: GetBalances :many
SELECT * FROM balances
WHERE user_id = @user_id;

-- name: CreateBalance :exec
INSERT INTO balances (user_id, asset, pending, sent, created_at, updated_at)
  VALUES(@user_id, @asset, 0, 0, now(), now());

-- name: UpdateBalancePending :exec
UPDATE balances
SET pending = @pending
WHERE user_id = @user_id
  AND asset = @asset;

-- name: UpdateBalanceSent :exec
UPDATE balances
SET sent = @sent
WHERE user_id = @user_id
  AND asset = @asset;

-- name: GetState :one
SELECT * FROM states
WHERE user_id = @user_id
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateState :exec
INSERT INTO states (user_id, state, data, meta, created_at)
  VALUES (@user_id, @state, @data, @meta, now());

-- name: UpdateStateData :exec
UPDATE states
SET data = @data
WHERE user_id = @user_id;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE user_id = @user_id
LIMIT 1;

-- name: CreateAccount :one
INSERT INTO accounts (user_id, username, address, seed, created_at)
  VALUES (@user_id, @username, @address, @seed, now()) RETURNING id;

-- name: GetAccountByKey :one
SELECT * FROM accounts
WHERE username = @key or address = @key;

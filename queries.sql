-- name: GetState :one
SELECT * FROM states
WHERE user_id = @user_id
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateState :exec
INSERT INTO states (user_id, state, data, meta, created_at)
  VALUES (@user_id, @state, @data, @meta, now());


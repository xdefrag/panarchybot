// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package db

import (
	"context"
)

const createState = `-- name: CreateState :exec
INSERT INTO states (user_id, state, data, meta, created_at)
  VALUES ($1, $2, $3, $4, now())
`

type CreateStateParams struct {
	UserID int64
	State  string
	Data   map[string]interface{}
	Meta   map[string]interface{}
}

func (q *Queries) CreateState(ctx context.Context, arg CreateStateParams) error {
	_, err := q.db.Exec(ctx, createState,
		arg.UserID,
		arg.State,
		arg.Data,
		arg.Meta,
	)
	return err
}

const getState = `-- name: GetState :one
SELECT user_id, state, data, meta, created_at FROM states
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 1
`

func (q *Queries) GetState(ctx context.Context, userID int64) (State, error) {
	row := q.db.QueryRow(ctx, getState, userID)
	var i State
	err := row.Scan(
		&i.UserID,
		&i.State,
		&i.Data,
		&i.Meta,
		&i.CreatedAt,
	)
	return i, err
}

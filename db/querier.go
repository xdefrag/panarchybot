// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"context"
)

type Querier interface {
	CreateAccount(ctx context.Context, arg CreateAccountParams) (int64, error)
	CreateState(ctx context.Context, arg CreateStateParams) error
	GetAccount(ctx context.Context, userID int64) (Account, error)
	GetAccountByKey(ctx context.Context, key string) (Account, error)
	GetState(ctx context.Context, userID int64) (State, error)
	UpdateStateData(ctx context.Context, arg UpdateStateDataParams) error
}

var _ Querier = (*Queries)(nil)

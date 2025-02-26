package db

import (
	"context"
	"time"
)

type QuerierWithTimeout struct {
	q       Querier
	timeout time.Duration
}

func WithTimeout(q Querier, timeout time.Duration) *QuerierWithTimeout {
	return &QuerierWithTimeout{
		q:       q,
		timeout: timeout,
	}
}

func (q *QuerierWithTimeout) CreateAccount(ctx context.Context, arg CreateAccountParams) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, q.timeout)
	defer cancel()
	return q.q.CreateAccount(ctx, arg)
}

func (q *QuerierWithTimeout) CreateState(ctx context.Context, arg CreateStateParams) error {
	ctx, cancel := context.WithTimeout(ctx, q.timeout)
	defer cancel()
	return q.q.CreateState(ctx, arg)
}

func (q *QuerierWithTimeout) GetAccount(ctx context.Context, userID int64) (Account, error) {
	ctx, cancel := context.WithTimeout(ctx, q.timeout)
	defer cancel()
	return q.q.GetAccount(ctx, userID)
}

func (q *QuerierWithTimeout) GetAccountByKey(ctx context.Context, key string) (Account, error) {
	ctx, cancel := context.WithTimeout(ctx, q.timeout)
	defer cancel()
	return q.q.GetAccountByKey(ctx, key)
}

func (q *QuerierWithTimeout) GetState(ctx context.Context, userID int64) (State, error) {
	ctx, cancel := context.WithTimeout(ctx, q.timeout)
	defer cancel()
	return q.q.GetState(ctx, userID)
}

func (q *QuerierWithTimeout) UpdateStateData(ctx context.Context, arg UpdateStateDataParams) error {
	ctx, cancel := context.WithTimeout(ctx, q.timeout)
	defer cancel()
	return q.q.UpdateStateData(ctx, arg)
}

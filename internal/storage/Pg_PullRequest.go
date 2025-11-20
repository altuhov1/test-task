package storage

import "github.com/jackc/pgx/v5/pgxpool"

type PullReqPostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPullReqPostgresStorage(pool *pgxpool.Pool) *PullReqPostgresStorage {

	return &PullReqPostgresStorage{pool: pool}
}

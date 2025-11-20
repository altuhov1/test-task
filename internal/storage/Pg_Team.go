package storage

import "github.com/jackc/pgx/v5/pgxpool"

type TeamPostgresStorage struct {
	pool *pgxpool.Pool
}

func NewTeamPostgresStorage(pool *pgxpool.Pool) *TeamPostgresStorage {

	return &TeamPostgresStorage{pool: pool}
}

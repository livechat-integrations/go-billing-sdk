//go:generate go run -mod=readonly github.com/sqlc-dev/sqlc/cmd/sqlc@v1.25.0 -x -f ./sqlc.yaml generate

package db

import "github.com/jackc/pgx/v5/pgxpool"

type DB struct {
	*pgxpool.Pool
}

func NewDB(pool *pgxpool.Pool) *Queries {
	return New(&DB{pool})
}

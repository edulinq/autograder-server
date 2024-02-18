package pg

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/edulinq/autograder/config"
)

type backend struct {
    pool *pgxpool.Pool
}

func Open() (*backend, error) {
    uri := config.DB_PG_URI.Get();
    if (uri == "") {
        return nil, fmt.Errorf("Postgres connection URI is empty.");
    }

    pool, err := pgxpool.New(context.Background(), uri);
	if (err != nil) {
        return nil, fmt.Errorf("Failed to open connection pool to Postgres database at '%s': %w.", uri, err);
	}

    return &backend{pool}, nil;
}

func (this *backend) Close() error {
    this.pool.Close()
    return nil;
}

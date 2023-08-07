package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Karzoug/goph_keeper/pkg/e"
	sconfig "github.com/Karzoug/goph_keeper/server/internal/config/storage"
)

const (
	URIPreffix            = "postgres:"
	prepareDBTimeout      = 5 * time.Second
	duplicateKeyErrorCode = "23505"
)

type storage struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, cfg sconfig.Config) (*storage, error) {
	op := "create postgres storage"

	ctx, cancel := context.WithTimeout(ctx, prepareDBTimeout)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.URI)
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, e.Wrap(op, err)
	}

	return &storage{
		db: pool,
	}, nil
}

func (s *storage) Close() error {
	const op = "postgres: close"

	s.db.Close()

	return nil
}

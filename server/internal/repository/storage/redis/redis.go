package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Karzoug/goph_keeper/pkg/e"
	sconfig "github.com/Karzoug/goph_keeper/server/internal/config/storage"
	"github.com/Karzoug/goph_keeper/server/internal/repository/storage"
)

const (
	URIPreffix = "redis:"
)

type client struct {
	rdb *redis.Client
}

// New creates a new redis client.
func New(cfg sconfig.Config) (*client, error) {
	const op = "create redis storage"

	opt, err := redis.ParseURL(cfg.URI)
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	return &client{
		rdb: redis.NewClient(opt),
	}, nil
}

// Get returns value by key.
func (q *client) Get(ctx context.Context, key string) (string, error) {
	const op = "redis: get"

	val, err := q.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", e.Wrap(op, storage.ErrRecordNotFound)
	}
	return val, e.Wrap(op, err)
}

// Set sets value by key.
func (q *client) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	const op = "redis: set"

	return e.Wrap(op, q.rdb.Set(ctx, key, value, expiration).Err())
}

// Delete deletes value by key.
func (q *client) Delete(ctx context.Context, key string) error {
	const op = "redis: delete"

	val, err := q.rdb.Del(ctx, key).Result()
	if val == 0 {
		return e.Wrap(op, storage.ErrNoRecordsAffected)
	}
	return e.Wrap(op, err)
}

// Close closes the redis client, releasing any open resources.
func (q *client) Close() error {
	const op = "redis: close"

	return e.Wrap(op, q.rdb.Close())
}

package service

import (
	"context"
	"time"

	"golang.org/x/exp/slog"

	"github.com/Karzoug/goph_keeper/server/internal/model/user"
)

type Storage interface {
	AddUser(context.Context, user.User) error
	GetUser(ctx context.Context, email string) (user.User, error)
	UpdateUser(context.Context, user.User) error
}

type kvStorage interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

type Option func(*Service)

type caches struct {
	auth kvStorage
}

type Service struct {
	cfg     Config
	storage Storage
	caches  caches
	logger  *slog.Logger
}

func New(cfg Config, storage Storage, logger *slog.Logger, options ...Option) (*Service, error) {
	const op = "service: create service"

	s := &Service{
		cfg:     cfg,
		storage: storage,
		logger:  logger,
	}

	for _, opt := range options {
		opt(s)
	}

	if s.caches.auth == nil {
		// TODO: set default cache
	}

	return s, nil
}

func WithAuthCache(cache kvStorage) Option {
	return func(s *Service) {
		s.caches.auth = cache
	}
}

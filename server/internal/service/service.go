package service

import (
	"context"

	"golang.org/x/exp/slog"

	"github.com/Karzoug/goph_keeper/server/internal/model/user"
)

type Storage interface {
	AddUser(context.Context, user.User) error
	GetUser(ctx context.Context, email string) (user.User, error)
	UpdateUser(context.Context, user.User) error
}

type Service struct {
	cfg     Config
	storage Storage
	logger  *slog.Logger
}

func New(cfg Config, storage Storage, logger *slog.Logger) (*Service, error) {
	const op = "service: create service"

	s := &Service{
		cfg:     cfg,
		storage: storage,
		logger:  logger,
	}

	return s, nil
}

package service

import (
	"context"
	"os"
	"time"

	"golang.org/x/exp/slog"

	scfg "github.com/Karzoug/goph_keeper/server/internal/config/service"
	"github.com/Karzoug/goph_keeper/server/internal/model/user"
	"github.com/Karzoug/goph_keeper/server/internal/repository/mail"
	"github.com/Karzoug/goph_keeper/server/internal/repository/rtask"
	"github.com/Karzoug/goph_keeper/server/internal/repository/storage/smap"
)

type Storage interface {
	AddUser(context.Context, user.User) error
	GetUser(ctx context.Context, email string) (user.User, error)
	UpdateUser(context.Context, user.User) error
	Close() error
}

type kvStorage interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Close() error
}

type mailSender interface {
	Send(context.Context, *mail.Mail) error
	Validate(email string) error
}

type Option func(*Service)

type caches struct {
	auth kvStorage
	mail kvStorage
}

type Service struct {
	cfg         scfg.Config
	storage     Storage
	caches      caches
	rtaskClient rtask.Client
	mailSender  mailSender
	logger      *slog.Logger
}

func New(cfg scfg.Config,
	storage Storage,
	rtaskClient rtask.Client,
	mailSender mailSender,
	options ...Option) (*Service, error) {
	s := &Service{
		cfg:         cfg,
		storage:     storage,
		rtaskClient: rtaskClient,
		mailSender:  mailSender,
	}

	for _, opt := range options {
		opt(s)
	}

	if s.logger == nil {
		s.logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	if s.caches.auth == nil {
		s.caches.auth = smap.New(30 * time.Minute)
	}
	if s.caches.mail == nil {
		s.caches.mail = smap.New(30 * time.Minute)
	}

	s.logger = s.logger.With("from", "service")

	return s, nil
}

func WithSLogger(logger *slog.Logger) Option {
	return func(s *Service) {
		s.logger = logger
	}
}

func WithAuthCache(cache kvStorage) Option {
	return func(s *Service) {
		s.caches.auth = cache
	}
}

func WithMailCache(cache kvStorage) Option {
	return func(s *Service) {
		s.caches.mail = cache
	}
}

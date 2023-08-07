package app

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"

	"github.com/Karzoug/goph_keeper/pkg/e"
	"github.com/Karzoug/goph_keeper/server/internal/config"
	scfg "github.com/Karzoug/goph_keeper/server/internal/config/service"
	"github.com/Karzoug/goph_keeper/server/internal/config/storage"
	"github.com/Karzoug/goph_keeper/server/internal/delivery/grpc"
	rtasks "github.com/Karzoug/goph_keeper/server/internal/delivery/rtask"
	"github.com/Karzoug/goph_keeper/server/internal/repository/mail/smtp"
	rtaskc "github.com/Karzoug/goph_keeper/server/internal/repository/rtask"
	"github.com/Karzoug/goph_keeper/server/internal/repository/storage/postgres"
	"github.com/Karzoug/goph_keeper/server/internal/repository/storage/redis"
	"github.com/Karzoug/goph_keeper/server/internal/repository/storage/sqlite"
	"github.com/Karzoug/goph_keeper/server/internal/service"
)

func Run(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	const op = "app run"

	serviceStorage, err := buildServiceStorage(ctx, cfg.Service.Storage)
	if err != nil {
		return e.Wrap(op, err)
	}
	defer serviceStorage.Close()
	logger.Info("app run: service storage created")

	rtaskClient, err := rtaskc.New(cfg.RTask.Storage.URI, logger)
	if err != nil {
		return e.Wrap(op, err)
	}
	logger.Info("app run: client for redis task manager created")

	smtpClient, err := smtp.New(cfg.SMTP)
	if err != nil {
		return e.Wrap(op, err)
	}
	logger.Info("app run: smtp client created")

	opts, err := buildServiceOptions(cfg.Service)
	if err != nil {
		return e.Wrap(op, err)
	}
	opts = append(opts, service.WithSLogger(logger))

	service, err := service.New(cfg.Service, serviceStorage, rtaskClient, smtpClient, opts...)
	if err != nil {
		return e.Wrap(op, err)
	}
	logger.Info("app run: service created")

	grpcServer, err := grpc.New(cfg.GRPC, service, logger)
	if err != nil {
		return e.Wrap(op, err)
	}
	logger.Info("app run: grpc server created")

	rtaskServer, err := rtasks.New(cfg.RTask, service, logger)
	if err != nil {
		return e.Wrap(op, err)
	}
	logger.Info("app run: server for redis task manager created")

	g := new(errgroup.Group)

	g.Go(func() error {
		return grpcServer.Run(ctx)
	})

	g.Go(func() error {
		return rtaskServer.Run(ctx)
	})

	if err := g.Wait(); err != nil {
		return e.Wrap(op, err)
	}

	return nil
}

func buildServiceStorage(ctx context.Context, cfg storage.Config) (service.Storage, error) {
	switch {
	case strings.HasPrefix(cfg.URI, postgres.URIPreffix):
		return postgres.New(ctx, cfg)
	case strings.HasPrefix(cfg.URI, sqlite.URIPreffix):
		return sqlite.New(cfg)
	case strings.HasPrefix(cfg.URI, "grpc://"):
		panic("not implemented")
	default:
		return nil, errors.New("unknown storage type")
	}
}

func buildServiceOptions(cfg scfg.Config) ([]service.Option, error) {
	opts := make([]service.Option, 0)

	if len(cfg.AuthCache.URI) != 0 {
		ac, err := buildServiceCache(cfg.AuthCache)
		if err != nil {
			return nil, err
		} else {
			opts = append(opts, service.WithAuthCache(ac))
		}
	}
	if len(cfg.MailCache.URI) != 0 {
		mc, err := buildServiceCache(cfg.MailCache)
		if err != nil {
			return nil, err
		} else {
			opts = append(opts, service.WithMailCache(mc))
		}
	}

	return opts, nil
}

func buildServiceCache(cfg storage.Config) (service.KvStorage, error) {
	switch {
	case strings.HasPrefix(cfg.URI, redis.URIPreffix):
		return redis.New(cfg)
	default:
		return nil, errors.New("unknown storage type")
	}
}

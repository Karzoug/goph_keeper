package rtask

import (
	"context"
	"errors"

	"github.com/hibiken/asynq"
	"golang.org/x/exp/slog"

	"github.com/Karzoug/goph_keeper/pkg/e"
	"github.com/Karzoug/goph_keeper/server/internal/config/rtask"
	"github.com/Karzoug/goph_keeper/server/internal/service"
	"github.com/Karzoug/goph_keeper/server/internal/service/task"
)

var errUnsupportedStorageType = errors.New("unsupported storage type")

type server struct {
	logger  *slog.Logger
	service *service.Service

	asynqServer *asynq.Server
}

func New(cfg rtask.Config, service *service.Service, logger *slog.Logger) (*server, error) {
	const op = "create rtask server"

	opt, err := asynq.ParseRedisURI(cfg.Storage.URI)
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	asynqServer := asynq.NewServer(opt,
		asynq.Config{
			Concurrency: cfg.Concurrency,
			LogLevel:    asynq.InfoLevel,
		},
	)

	srv := &server{
		logger:      logger.With("from", "rtask server"),
		service:     service,
		asynqServer: asynqServer,
	}

	return srv, nil
}

func (s *server) Run(ctx context.Context) error {
	const op = "rtask server: run"

	s.logger.Info("running")

	mux := asynq.NewServeMux()
	mux.HandleFunc(task.TypeWelcomeVerificationEmail, s.service.HandleWelcomeVerificationEmailTask)

	idleConnsClosed := make(chan struct{})

	go func() {
		<-ctx.Done()
		s.shutdown()
		close(idleConnsClosed)
	}()

	if err := s.asynqServer.Start(mux); err != nil {
		return e.Wrap(op, err)
	}

	<-idleConnsClosed

	return nil
}

func (s *server) shutdown() {
	s.logger.Info("shutting down")

	s.asynqServer.Shutdown()
}

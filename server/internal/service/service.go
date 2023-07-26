package service

import "golang.org/x/exp/slog"

type Service struct {
	cfg    Config
	logger *slog.Logger
}

func New(cfg Config, logger *slog.Logger) (*Service, error) {
	const op = "service: create service"

	s := &Service{
		logger: logger,
		cfg:    cfg,
	}

	return s, nil
}

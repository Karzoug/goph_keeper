package grpc

import (
	"context"
	"net"

	"golang.org/x/exp/slog"
	"google.golang.org/grpc"

	pb "github.com/Karzoug/goph_keeper/common/grpc"
	"github.com/Karzoug/goph_keeper/pkg/e"
	gcfg "github.com/Karzoug/goph_keeper/server/internal/config/grpc"
	"github.com/Karzoug/goph_keeper/server/internal/delivery/grpc/interceptor/auth"
	"github.com/Karzoug/goph_keeper/server/internal/service"
)

type server struct {
	cfg     gcfg.Config
	logger  *slog.Logger
	service *service.Service

	grpcServer *grpc.Server
	pb.UnimplementedGophKeeperServiceServer
}

func New(cfg gcfg.Config, service *service.Service, logger *slog.Logger) *server {
	publicMethods := []string{
		pb.GophKeeperService_Register_FullMethodName,
		pb.GophKeeperService_Login_FullMethodName,
	}

	// add TLS
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			auth.AuthUnaryServerInterceptor(service.AuthUser, publicMethods, logger)))

	ss := &server{
		cfg:        cfg,
		logger:     logger.With("from", "grpc server"),
		service:    service,
		grpcServer: grpcServer,
	}

	pb.RegisterGophKeeperServiceServer(ss.grpcServer, ss)

	return ss
}

func (s *server) Run(ctx context.Context) error {
	const op = "run"

	s.logger.Info("running", slog.String("address", s.cfg.Address))

	idleConnsClosed := make(chan struct{})

	var lc net.ListenConfig
	listen, err := lc.Listen(ctx, "tcp", s.cfg.Address)
	if err != nil {
		return e.Wrap(op, err)
	}

	go func() {
		<-ctx.Done()
		s.shutdown()
		close(idleConnsClosed)
	}()

	if err := s.grpcServer.Serve(listen); err != nil {
		return e.Wrap(op, err)
	}

	<-idleConnsClosed

	return nil
}

func (s *server) shutdown() {
	s.logger.Info("shutting down")

	s.grpcServer.GracefulStop()
}

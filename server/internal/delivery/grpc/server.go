package grpc

import (
	"context"
	"crypto/tls"
	"net"

	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/Karzoug/goph_keeper/common/grpc/server"
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
	pb.UnimplementedGophKeeperServerServer
}

func New(cfg gcfg.Config, service *service.Service, logger *slog.Logger) (*server, error) {
	const op = "create grpc server"

	publicMethods := []string{
		pb.GophKeeperServer_Register_FullMethodName,
		pb.GophKeeperServer_Login_FullMethodName,
	}

	tlsCfg, err := loadConfig(cfg.CertFileName, cfg.KeyFileName)
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	grpcServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsCfg)),
		grpc.UnaryInterceptor(
			auth.AuthUnaryServerInterceptor(service.AuthUser, publicMethods, logger)))

	ss := &server{
		cfg:        cfg,
		logger:     logger.With("from", "grpc server"),
		service:    service,
		grpcServer: grpcServer,
	}

	pb.RegisterGophKeeperServerServer(ss.grpcServer, ss)

	return ss, nil
}

func (s *server) Run(ctx context.Context) error {
	const op = "run"

	s.logger.Info("running", slog.String("address", s.cfg.Address()))

	idleConnsClosed := make(chan struct{})

	var lc net.ListenConfig
	listen, err := lc.Listen(ctx, "tcp", s.cfg.Address())
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

// loadConfig creates a new TLS config from the given certificate and key files.
func loadConfig(certFilename, keyFilename string) (*tls.Config, error) {
	serverCert, err := tls.LoadX509KeyPair(certFilename, keyFilename)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{serverCert},
	}, nil
}

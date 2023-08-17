package grpc

import (
	"context"
	"errors"

	pb "github.com/Karzoug/goph_keeper/common/grpc"
	"github.com/Karzoug/goph_keeper/pkg/logger/slog/sl"
	"github.com/Karzoug/goph_keeper/server/internal/service"
)

func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	const op = "register user"

	if err := s.service.Register(ctx, req.Email, req.Hash); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmailFormat):
			return nil, pb.ErrInvalidEmailFormat
		case errors.Is(err, service.ErrInvalidHashFormat):
			return nil, pb.ErrInvalidHashFormat
		case errors.Is(err, service.ErrUserAlreadyExists):
			return nil, pb.ErrUserAlreadyExists
		default:
			s.logger.Error(op, sl.Error(err))
			return nil, pb.ErrInternal
		}
	}

	return &pb.RegisterResponse{}, nil
}

func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	const op = "login user"

	var (
		token string
		err   error
	)
	if req.EmailCode != "" {
		token, err = s.service.LoginWithEmailCode(ctx, req.Email, req.Hash, req.EmailCode)
	} else {
		token, err = s.service.Login(ctx, req.Email, req.Hash)
	}

	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmailFormat):
			return nil, pb.ErrInvalidEmailFormat
		case errors.Is(err, service.ErrInvalidHashFormat):
			return nil, pb.ErrInvalidHashFormat
		case errors.Is(err, service.ErrUserInvalidHash):
			return nil, pb.ErrUserInvalidHash
		case errors.Is(err, service.ErrUserEmailNotVerified):
			return nil, pb.ErrUserEmailNotVerified
		case errors.Is(err, service.ErrUserNotExists):
			return nil, pb.ErrUserNotExists
		default:
			s.logger.Error(op, sl.Error(err))
			return nil, pb.ErrInternal
		}
	}

	return &pb.LoginResponse{Token: token}, nil
}

package grpc

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInternal               = status.Error(codes.Internal, "internal errors")
	ErrUserAlreadyExists      = status.Error(codes.AlreadyExists, "user already exists")
	ErrUserNotExists          = status.Error(codes.NotFound, "user not exists")
	ErrUserEmailNotVerified   = status.Error(codes.Unauthenticated, "user email not verified")
	ErrUserInvalidHash        = status.Error(codes.Unauthenticated, "user hash not valid")
	ErrUserNeedAuthentication = status.Error(codes.Unauthenticated, "user need authentication")
	ErrUserInvalidToken       = status.Error(codes.InvalidArgument, "user invalid token")
	ErrInvalidEmailFormat     = status.Error(codes.InvalidArgument, "invalid email format")
	ErrInvalidHashFormat      = status.Error(codes.InvalidArgument, "invalid hash format")
	ErrEmptyAuthData          = status.Error(codes.InvalidArgument, "empty auth data")
)

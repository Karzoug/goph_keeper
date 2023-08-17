package auth

import (
	"context"
	"errors"

	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	gerr "github.com/Karzoug/goph_keeper/common/grpc"
	"github.com/Karzoug/goph_keeper/pkg/logger/slog/sl"
	"github.com/Karzoug/goph_keeper/server/internal/service"
)

type (
	authContextKey int8
	AuthFunc       func(ctx context.Context, token string) (string, error)
)

const emailAuthCtxKey authContextKey = 0

var ErrCtxEmailNotFound = errors.New("email not found in context")

func AuthUnaryServerInterceptor(authFunc AuthFunc, publicMethods []string, logger *slog.Logger) grpc.UnaryServerInterceptor {
	isPublicMethodCheckFnc := func(m string) bool {
		for i := 0; i < len(publicMethods); i++ {
			if m == publicMethods[i] {
				return true
			}
		}
		return false
	}

	logger = logger.With("from", "auth grpc interceptor")

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if isPublicMethodCheckFnc(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, gerr.ErrEmptyAuthData
		}

		tokenSlice := md["token"]
		if len(tokenSlice) == 0 {
			return nil, gerr.ErrEmptyAuthData
		}

		email, err := authFunc(ctx, tokenSlice[0])
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidTokenFormat):
				return nil, gerr.ErrInvalidTokenFormat
			case errors.Is(err, service.ErrUserNeedAuthentication):
				return nil, gerr.ErrUserNeedAuthentication
			default:
				logger.Error("auth user failed", sl.Error(err))
				return nil, gerr.ErrInternal
			}
		}

		newCtx := context.WithValue(ctx, emailAuthCtxKey, email)
		return handler(newCtx, req)
	}
}

func EmailFromContext(ctx context.Context) (string, error) {
	value := ctx.Value(emailAuthCtxKey)
	if value == nil {
		return "", ErrCtxEmailNotFound
	}

	return value.(string), nil
}

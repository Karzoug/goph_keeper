package server

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrInternal returned if server for some reason cannot process the request and there is no more appropriate error.
	ErrInternal = status.Error(codes.Internal, "internal errors")
	// ErrUserAlreadyExists returned on registration when the user already exists.
	ErrUserAlreadyExists = status.Error(codes.AlreadyExists, "user already exists")
	// ErrUserNotExists returned on login when the user does not exist.
	ErrUserNotExists = status.Error(codes.NotFound, "user not exists")
	// ErrUserEmailNotVerified returned on login when the user email is not yet verified.
	ErrUserEmailNotVerified = status.Error(codes.Unauthenticated, "user email not verified")
	// ErrUserInvalidHash returned if the passed authentication hash is not valid i.e. does not match the user.
	// See also ErrInvalidHashFormat description.
	ErrUserInvalidHash = status.Error(codes.Unauthenticated, "user hash not valid")
	// ErrUserNeedAuthentication returned if user token is no longer valid (expired for example).
	ErrUserNeedAuthentication = status.Error(codes.Unauthenticated, "user need authentication")
	// ErrInvalidTokenFormat returned if user send token with invalid format.
	ErrInvalidTokenFormat = status.Error(codes.InvalidArgument, "user invalid token format")
	// ErrInvalidEmailFormat returned if format of the passed email is not valid.
	ErrInvalidEmailFormat = status.Error(codes.InvalidArgument, "invalid email format")
	// ErrInvalidHashFormat returned if format of the passed authentication hash is not valid.
	// See also ErrUserInvalidHash description.
	ErrInvalidHashFormat = status.Error(codes.InvalidArgument, "invalid hash format")
	// ErrEmptyAuthData returned if no auth data is passed.
	ErrEmptyAuthData = status.Error(codes.InvalidArgument, "empty auth data")
	// ErrVaultItemVersionConflict returned if the client has changed the item and is trying to send it to the server,
	// but the item on the server has changed since the last synchronization.
	ErrVaultItemConflictVersion = status.Error(codes.InvalidArgument, "vault item: conflict version")
	// ErrVaultItemValueTooBig returned if the client is trying to send large data using an inappropriate method.
	ErrVaultItemValueTooBig = status.Error(codes.OutOfRange, "vault item: big value")
	// ErrLimitVaultSizeExceeded returned if the client vault is full, the limit is exceeded.
	ErrLimitVaultSizeExceeded = status.Error(codes.ResourceExhausted, "limit vault size exceeded")
	// ErrLimitFileSizeExceeded returned if the file size sent by the user is too big, the limit is exceeded.
	ErrLimitFileSizeExceeded = status.Error(codes.OutOfRange, "limit file size exceeded")
	// ErrFileNotFound returned if the file is not found.
	ErrFileNotFound = status.Error(codes.NotFound, "file not found")
)

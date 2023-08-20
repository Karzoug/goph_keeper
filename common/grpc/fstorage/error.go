package fstorage

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrInternal returned if server for some reason cannot process the request and there is no more appropriate error.
	ErrInternal = status.Error(codes.Internal, "internal errors")
	// ErrLimitVaultSizeExceeded returned if the client vault is full, the limit is exceeded.
	ErrLimitVaultSizeExceeded = status.Error(codes.ResourceExhausted, "limit vault size exceeded")
	// ErrLimitFileSizeExceeded returned if the file size sent by the user is too big, the limit is exceeded.
	ErrLimitFileSizeExceeded = status.Error(codes.OutOfRange, "limit file size exceeded")
	// ErrFileNotFound returned if the file is not found.
	ErrFileNotFound = status.Error(codes.NotFound, "file not found")
)

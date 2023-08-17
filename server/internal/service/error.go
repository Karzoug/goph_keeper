package service

import "errors"

var (
	ErrUserAlreadyExists        = errors.New("user already exists")
	ErrUserNotExists            = errors.New("user not exists")
	ErrUserEmailNotVerified     = errors.New("user email not verified")
	ErrUserInvalidHash          = errors.New("user hash not valid")
	ErrInvalidEmailFormat       = errors.New("invalid email format")
	ErrInvalidHashFormat        = errors.New("invalid hash format")
	ErrInvalidTokenFormat       = errors.New("user token invalid format")
	ErrUserNeedAuthentication   = errors.New("user need authentication")
	ErrVaultItemVersionConflict = errors.New("vault item: conflict version")
	ErrVaultItemValueTooBig     = errors.New("vault item: big value")
)

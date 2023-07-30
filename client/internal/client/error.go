package client

import (
	"errors"
	"fmt"
)

var (
	ErrPasswordTooShort             = fmt.Errorf("password too short (must be at least %d characters)", MinPasswordLength)
	ErrInvalidEmail                 = errors.New("invalid email")
	ErrUserAlreadyExists            = errors.New("user already exists")
	ErrUserInvalidPassword          = errors.New("invalid password")
	ErrUserEmailNotVerified         = errors.New("email not verified")
	ErrInvalidEmailVerificationCode = errors.New("invalid email verification code")
	ErrUserNotExists                = errors.New("user not exists")
	ErrAppInternal                  = errors.New("app internal error")
	ErrServerInternal               = errors.New("server internal error")
	ErrUserNeedAuthentication       = errors.New("user need authentication")
	errUnresolvableHost             = errors.New("unresolvable host")
)

package client

import (
	"context"
	"errors"
	"net"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/idna"

	"github.com/Karzoug/goph_keeper/client/internal/model/auth"
	"github.com/Karzoug/goph_keeper/client/pkg/crypto"
	pb "github.com/Karzoug/goph_keeper/common/grpc"
)

const MinPasswordLength = 8

// Register registers a new user on the server with the specified email and password.
// (!) Method wipes the password in memory to prevent long-term storage.
// Method returns an error if the email or password is not valid.
func (c *Client) Register(ctx context.Context, email string, password []byte) error {
	const op = "register user"

	defer crypto.Wipe(password) // prevent long-term storage of the password in memory

	if len(email) == 0 {
		return ErrInvalidEmail
	}

	if err := validateMX(email); err != nil {
		return ErrInvalidEmail
	}

	if utf8.RuneCount(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	hash, err := auth.NewHash([]byte(email), password)
	if err != nil {
		if errors.Is(err, auth.ErrEmptyPassword) {
			return ErrPasswordTooShort
		}
		c.logger.Debug(op, err)
		return ErrAppInternal
	}

	_, err = c.grpcClient.Register(ctx, &pb.RegisterRequest{
		Email: email,
		Hash:  hash,
	})
	if err != nil {
		switch {
		case errors.Is(err, pb.ErrInvalidEmailFormat):
			return ErrInvalidEmail
		case errors.Is(err, pb.ErrUserAlreadyExists):
			return ErrUserAlreadyExists
		default:
			c.logger.Debug(op, err)
			return ErrServerInternal
		}
	}

	return nil
}

func (c *Client) Login(ctx context.Context, email string, password []byte) error {
	const op = "login user"

	if len(email) == 0 {
		return ErrInvalidEmail
	}

	if utf8.RuneCount(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	if err := c.setPasswordHashes(ctx, email, password); err != nil {
		c.logger.Debug(op, err)
		return ErrAppInternal
	}

	resp, err := c.grpcClient.Login(ctx, &pb.LoginRequest{
		Email: c.credentials.email,
		Hash:  c.credentials.authHash,
	})
	if err != nil {
		switch {
		case errors.Is(err, pb.ErrUserInvalidHash):
			_ = c.clearCredentials(ctx)
			return ErrUserInvalidPassword
		case errors.Is(err, pb.ErrUserEmailNotVerified):
			return ErrUserEmailNotVerified
		case errors.Is(err, pb.ErrUserNotExists):
			_ = c.clearCredentials(ctx)
			return ErrUserNotExists
		default:
			c.logger.Debug(op, err)
			return ErrServerInternal
		}
	}

	if err := c.setToken(resp.Token); err != nil {
		c.logger.Debug(op, err)
		return ErrAppInternal
	}
	return nil
}

func (c *Client) VerifyEmail(ctx context.Context, code string) error {
	const op = "verify email"

	resp, err := c.grpcClient.Login(ctx, &pb.LoginRequest{
		Email:     c.credentials.email,
		Hash:      c.credentials.authHash,
		EmailCode: code,
	})
	if err != nil {
		switch {
		case errors.Is(err, pb.ErrUserInvalidHash):
			_ = c.clearCredentials(ctx)
			return ErrUserInvalidPassword
		case errors.Is(err, pb.ErrUserEmailNotVerified):
			return ErrInvalidEmailVerificationCode
		case errors.Is(err, pb.ErrUserNotExists):
			_ = c.clearCredentials(ctx)
			return ErrUserNotExists
		default:
			c.logger.Debug(op, err)
			return ErrServerInternal
		}
	}

	if err := c.setToken(resp.Token); err != nil {
		c.logger.Debug(op, err)
		return ErrAppInternal
	}
	return nil
}

func (c *Client) Logout(ctx context.Context) error {
	const op = "logout user"

	if err := c.credentialsStorage.DeleteCredentials(); err != nil {
		c.logger.Debug(op, err)
		return ErrAppInternal
	}

	return nil
}

// validateMX validate if MX record exists for a domain.
func validateMX(email string) error {
	_, host := split(email)
	if len(host) == 0 {
		return errUnresolvableHost
	}
	host = hostToASCII(host)
	if _, err := net.LookupMX(host); err != nil {
		return errUnresolvableHost
	}

	return nil
}

func split(email string) (account, host string) {
	i := strings.LastIndexByte(email, '@')
	// If no @ present, not a valid email.
	if i < 0 {
		return
	}
	account = email[:i]
	host = email[i+1:]
	return
}

// domainToASCII converts any internationalized domain names to ASCII
// reference: https://en.wikipedia.org/wiki/Punycode
func hostToASCII(host string) string {
	asciiDomain, err := idna.ToASCII(host)
	if err != nil {
		return host
	}
	return asciiDomain
}

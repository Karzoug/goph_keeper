package client

import (
	"context"
	"errors"
	"net/mail"
	"unicode/utf8"

	"github.com/Karzoug/goph_keeper/client/internal/model/auth"
	"github.com/Karzoug/goph_keeper/client/pkg/crypto"
	pb "github.com/Karzoug/goph_keeper/common/grpc"
)

const MinPasswordLength = 8

// Register registers a new user on the server with the gieven email and password.
// Method returns an error if the email or password is not valid.
//
// Warning(!): method wipes the given password slice to prevent long-term storage in memory.
func (c *Client) Register(ctx context.Context, email string, password []byte) error {
	const op = "register user"

	defer crypto.Wipe(password) // prevent long-term storage of the password in memory

	if !isValidEmail(email) {
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

// Login builds local credentials. Then connects to the server:
//
// 1. connection error: if local vault owner email is equal to the given email,
// saves the local credentials, application works offline,
// otherwise it returns ErrServerInternal;
//
// 2. on server authentication failure:
// does not save data, returns ErrUserNeedAuthentication;
//
// 3. case of unverified mail: returns ErrUserEmailNotVerified;
//
// 4. on success: saves the data and the received token,
// if local vault owner email is not equal to the given email,
// then the local vault will be cleared.
func (c *Client) Login(ctx context.Context, email string, password []byte) error {
	const op = "login user"

	if !isValidEmail(email) {
		return ErrInvalidEmail
	}

	if utf8.RuneCount(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	hash, encrKey, err := buildPasswordHashes(ctx, email, password)
	if err != nil {
		c.logger.Debug(op, err)
		return ErrAppInternal
	}

	resp, err := c.grpcClient.Login(ctx, &pb.LoginRequest{
		Email: email,
		Hash:  []byte(hash),
	})
	if err != nil {
		switch {
		case errors.Is(err, pb.ErrUserInvalidHash):
			_ = c.clearCredentials(ctx)
			return ErrUserInvalidPassword
		case errors.Is(err, pb.ErrUserEmailNotVerified):
			if err := c.setCredentialsForced(ctx, email, hash, encrKey); err != nil {
				c.logger.Debug(op, err)
				return ErrAppInternal
			}
			return ErrUserEmailNotVerified
		case errors.Is(err, pb.ErrUserNotExists):
			_ = c.clearCredentials(ctx)
			return ErrUserNotExists
		default:
			// problems with grpc,
			// but if this is the owner of the vault, then they can try to work offline
			if err := c.setCredentialsForOwnerOnly(ctx, email, hash, encrKey); err != nil {
				c.logger.Debug(op, err)
				if errors.Is(err, ErrUserNeedAuthentication) {
					return ErrServerInternal
				}
				return ErrAppInternal
			}
			c.logger.Debug(op, err)
			return nil
		}
	}

	if err := c.setCredentialsForced(ctx, email, hash, encrKey); err != nil {
		c.logger.Debug(op, err)
		return ErrAppInternal
	}
	if err := c.setToken(resp.Token); err != nil {
		c.logger.Debug(op, err)
		return ErrAppInternal
	}
	return nil
}

// VerifyEmail sends a verification code to the server and
// returns nil only if successful.
func (c *Client) VerifyEmail(ctx context.Context, code string) error {
	const op = "verify email"

	if len(c.credentials.email) == 0 ||
		c.credentials.authHash == nil {
		return ErrUserNeedAuthentication
	}

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

func isValidEmail(email string) bool {
	if len(email) == 0 {
		return false
	}
	e, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	email = e.Address

	return true
}

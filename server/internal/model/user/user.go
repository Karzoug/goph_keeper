package user

import (
	"errors"
	"net"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/net/idna"

	"github.com/Karzoug/goph_keeper/server/internal/model/auth"
)

var (
	ErrInvalidHashFormat  = errors.New("invalid hash format")
	ErrInvalidEmailFormat = errors.New("invalid email format")
	errUnresolvableHost   = errors.New("unresolvable host")
)

// User is a service user (client).
type User struct {
	Email           string
	IsEmailVerified bool
	AuthKey         auth.Key
	CreatedAt       time.Time
}

// New returns a new user.
func New(email string, authHash []byte) (User, error) {
	if !isValidEmail(email) {
		return User{}, ErrInvalidEmailFormat
	}

	authKey, err := auth.NewKey(authHash)
	if err != nil {
		return User{}, ErrInvalidHashFormat
	}

	return User{
		Email:           email,
		IsEmailVerified: false,
		AuthKey:         authKey,
		CreatedAt:       time.Now(),
	}, nil
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
	if err := validateMX(email); err != nil {
		return false
	}

	return true
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

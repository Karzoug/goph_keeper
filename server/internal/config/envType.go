package config

import (
	"errors"
	"strings"
)

const (
	// EnvDevelopment is a constant defining the development environment.
	EnvProduction EnvType = iota
	// EnvProduction is a constant defining the production environment.
	EnvDevelopment
)

const (
	envProductionString  = "production"
	envDevelopmentString = "development"
	envUnknownString     = "unknown"
)

// ErrUnknownEnv is an error returned when the environment is unknown.
var ErrUnknownEnv = errors.New("unknown environment mode")

// EnvType is a type of environment (production or development).
type EnvType int8

// String returns the string representation of the environment type variable.
func (e EnvType) String() string {
	switch e {
	case EnvDevelopment:
		return envDevelopmentString
	case EnvProduction:
		return envProductionString
	default:
		return envUnknownString
	}
}

// EnvTypeParserFunc is a function that parses the environment variable string.
var EnvTypeParserFunc = func(v string) (interface{}, error) {
	switch {
	case strings.HasPrefix(v, "dev"):
		return EnvDevelopment, nil
	case strings.HasPrefix(v, "prod"):
		return EnvProduction, nil
	}
	return nil, ErrUnknownEnv
}

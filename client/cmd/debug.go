//go:build debug

package main

import "github.com/Karzoug/goph_keeper/client/internal/config"

func init() {
	envMode = config.EnvDevelopment
}

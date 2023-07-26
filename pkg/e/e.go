// Package e provides a helper function to work with errors.
package e

import (
	"fmt"
)

// Wrap returns a new wrapped error: combination of msg and err.
//
// If err is not nil the returned error will implement an Unwrap method returning err.
// If err is nil the returned error will be nil.
func Wrap(msg string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", msg, err)
}

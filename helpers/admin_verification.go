package helpers

import (
	"fmt"
)

// VerifyResourceExists verifies a resource was retrieved from context.
// This is a generic helper used across admin step definitions.
func VerifyResourceExists[T any](value T, ok bool, errorPrefix string) error {
	if !ok {
		return fmt.Errorf("%s was not retrieved", errorPrefix)
	}

	// For pointer types, check for nil
	if interface{}(value) == nil {
		return fmt.Errorf("%s is nil", errorPrefix)
	}
	return nil
}

package helpers

import (
	"fmt"
	"reflect"
)

// VerifyResourceExists verifies a resource was retrieved from context.
// This is a generic helper used across admin step definitions.
func VerifyResourceExists[T any](value T, ok bool, errorPrefix string) error {
	if !ok {
		return fmt.Errorf("%s was not retrieved", errorPrefix)
	}

	// For pointer types, check for nil using reflection.
	// The interface{}(value) == nil check fails for nil pointers because
	// a nil pointer wrapped in interface{} preserves its type info.
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Pointer && v.IsNil() {
		return fmt.Errorf("%s is nil", errorPrefix)
	}
	return nil
}

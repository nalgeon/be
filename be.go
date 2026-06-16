// Package be provides minimal assertions for Go tests.
//
// It only has three functions: [Equal], [Err], and [True],
// which are perfectly enough to write good tests.
//
// Example usage:
//
//	func Test(t *testing.T) {
//		re, err := regexp.Compile("he??o")
//		be.Err(t, err, nil) // expects no error
//		be.True(t, strings.Contains(re.String(), "?"))
//		be.Equal(t, re.String(), "he??o")
//	}
package be

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// equaler is an interface for types with an Equal method
// (like time.Time or net.IP).
type equaler[T any] interface {
	Equal(T) bool
}

// Equal asserts that got is equal to any of the wanted values.
func Equal[T any](tb testing.TB, got T, wants ...T) {
	tb.Helper()

	if len(wants) == 0 {
		tb.Fatal("no wants given")
		return
	}

	// Check if got matches any of the wants.
	for _, want := range wants {
		if areEqual(got, want) {
			return
		}
	}

	// There are no matches, report the failure.
	if len(wants) == 1 {
		// There is only one want, report it directly.
		tb.Errorf("got: %#v; want: %#v", got, wants[0])
		return
	}
	// There are multiple wants, report a summary.
	tb.Errorf("got: %#v; want any of: %v", got, wants)
}

// Err asserts that the got error matches any of the wanted values.
// The matching logic depends on want:
//   - If want is nil, checks if got is nil.
//   - If want is a string, checks if got.Error() contains want.
//   - If want is an error, checks if its value is found
//     in the got's error tree using [errors.Is].
//   - If want is a [reflect.Type], checks if its type is found
//     in the got's error tree using [errors.As].
//   - Otherwise fails the check.
//
// If no wants are given, checks if got is not nil.
func Err(tb testing.TB, got error, wants ...any) {
	tb.Helper()

	// If no wants are given, we expect got to be a non-nil error.
	if len(wants) == 0 {
		if got == nil {
			tb.Error("got: <nil>; want: error")
		}
		return
	}

	// Special case: there's only one want, it's nil, but got is not nil.
	// This is a fatal error, so we fail the test immediately.
	if len(wants) == 1 && wants[0] == nil {
		if got != nil {
			tb.Fatalf("unexpected error: %v", got)
			return
		}
	}

	// Check if got matches any of the wants.
	var message string
	for _, want := range wants {
		errMsg := checkErr(got, want)
		if errMsg == "" {
			return
		}
		if message == "" {
			message = errMsg
		}
	}

	// There are no matches, report the failure.
	if len(wants) == 1 {
		// There is only one want, report it directly.
		tb.Error(message)
		return
	}
	// There are multiple wants, report a summary.
	tb.Errorf("got: %T(%v); want any of: %v", got, got, wants)
}

// True asserts that got is true.
func True(tb testing.TB, got bool) {
	tb.Helper()
	if !got {
		tb.Error("got: false; want: true")
	}
}

// areEqual checks if a and b are equal.
func areEqual[T any](a, b T) bool {
	// Check if both are nil.
	if isNil(a) && isNil(b) {
		return true
	}

	// Try to compare using an Equal method.
	if eq, ok := any(a).(equaler[T]); ok {
		return eq.Equal(b)
	}

	// Special case for byte slices.
	if aBytes, ok := any(a).([]byte); ok {
		bBytes := any(b).([]byte)
		return bytes.Equal(aBytes, bBytes)
	}

	// Fallback to reflective comparison.
	return reflect.DeepEqual(a, b)
}

// isNil checks if v is nil.
func isNil(v any) bool {
	if v == nil {
		return true
	}

	// A non-nil interface can still hold a nil value,
	// so we must check the underlying value.
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return rv.IsNil()
	default:
		return false
	}
}

// checkErr checks if got error matches the want value.
// Returns an empty string if there's a match.
// Otherwise, returns an error message.
func checkErr(got error, want any) string {
	if want != nil && got == nil {
		return "got: <nil>; want: error"
	}

	switch w := want.(type) {
	case nil:
		if got != nil {
			return fmt.Sprintf("unexpected error: %v", got)
		}
	case string:
		if !strings.Contains(got.Error(), w) {
			return fmt.Sprintf("got: %q; want: %q", got.Error(), w)
		}
	case error:
		if !errors.Is(got, w) {
			return fmt.Sprintf("got: %T(%v); want: %T(%v)", got, got, w, w)
		}
	case reflect.Type:
		target := reflect.New(w).Interface()
		if !errors.As(got, target) {
			return fmt.Sprintf("got: %T; want: %s", got, w)
		}
	default:
		return fmt.Sprintf("unsupported want type: %T", want)
	}
	return ""
}

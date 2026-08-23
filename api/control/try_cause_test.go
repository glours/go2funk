package control_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/glours/go2funk/api/control"
)

// A Failure must always carry the error that caused it, whatever produced it.
// Regression tests: TryOf, MapTry and FlatMapTry used to build a Failure with a
// nil cause, so OrElseCause returned (zero, nil) and callers reading err saw a
// success.

var errBoom = errors.New("boom")

// causeOf fails the test if try is not a Failure, and returns its cause.
func causeOf[A any](t *testing.T, try control.Try[A]) error {
	t.Helper()
	if !try.IsFailure() {
		t.Fatal("expected a Failure, got a Success")
	}
	_, err := try.OrElseCause()
	if err == nil {
		t.Fatal("a Failure must not report a nil cause")
	}
	return err
}

func TestTryOfKeepsTheLambdaError(t *testing.T) {
	cause := causeOf(t, control.TryOf(func() (int, error) { return 0, errBoom }))
	if !errors.Is(cause, errBoom) {
		t.Errorf("cause = %v, want %v", cause, errBoom)
	}
}

func TestMapTryKeepsTheCause(t *testing.T) {
	failure := control.FailureOf[int](errBoom)

	cause := causeOf(t, control.MapTry(failure, strconv.Itoa))
	if !errors.Is(cause, errBoom) {
		t.Errorf("cause = %v, want %v", cause, errBoom)
	}
}

func TestFlatMapTryKeepsTheCause(t *testing.T) {
	failure := control.FailureOf[int](errBoom)
	mapper := func(value int) control.Try[string] { return control.SuccessOf(strconv.Itoa(value)) }

	cause := causeOf(t, control.FlatMapTry(failure, mapper))
	if !errors.Is(cause, errBoom) {
		t.Errorf("cause = %v, want %v", cause, errBoom)
	}
}

func TestFlatMapTryKeepsTheMapperCause(t *testing.T) {
	errMapper := errors.New("mapper failed")
	success := control.SuccessOf(10)
	mapper := func(value int) control.Try[string] { return control.FailureOf[string](errMapper) }

	cause := causeOf(t, control.FlatMapTry(success, mapper))
	if !errors.Is(cause, errMapper) {
		t.Errorf("cause = %v, want %v", cause, errMapper)
	}
}

func TestMapTryOnSuccessStillMaps(t *testing.T) {
	mapped := control.MapTry(control.SuccessOf(10), strconv.Itoa)

	if mapped.IsFailure() {
		t.Fatal("mapping a Success must produce a Success")
	}
	if got := mapped.OrElse("none"); got != "10" {
		t.Errorf("value = %q, want %q", got, "10")
	}
}

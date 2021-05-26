// Package control provides control structures such as Option, Try or Either...
package control

import "errors"

// Try control type allows user to write code without focusing on error management.
// internal implementations of Try are Success and Failure.
type Try[A any] interface {
	IsFailure() bool
	OrElse(A) A
	OrElseCause() (A, error)
	Filter(func(A) bool, error) Try[A]
}

// MapTry maps the element of a Try[A] to a new Try with element of type B.
// the mapper function should take a A value and return a B value.
func MapTry[A, B any](try Try[A], mapper func(A) B) Try[B] {
	if try.IsFailure() {
		return Failure[B]{}
	}
	return Success[B]{mapper(try.OrElse(*new(A)))}
}

// FlatMapTry maps the element of a Try[A] to a new Try with element of type B.
// the mapper function should take a A value and return a Try[B] as result.
func FlatMapTry[A, B any](try Try[A], mapper func(A) Try[B]) Try[B] {
	if try.IsFailure() {
		return Failure[B]{}
	}
	return mapper(try.OrElse(*new(A)))
}

// TryOf returns a Try[A] depending of the execution result of the lambda passed as parameter.
func TryOf[A any](lambda func() (A, error)) Try[A] {
	value, err := lambda()
	if err != nil {
		return Failure[A]{}
	}
	return Success[A]{value}
}

// SuccessOf returns a Success[A] as Try[A].
func SuccessOf[A any](value A) Try[A] {
	return Success[A]{value}
}

// FailureOf returns a Failure[A] as Try[A].
func FailureOf[A any](cause error) Try[A] {
	return Failure[A]{cause}
}

// Success is an implementation of Try with a defined value.
type Success[A any] struct {
	value A
}

// IsFailure checks if the current Try is a failure or not.
// for the Success implementation the value return is false.
func (s Success[A]) IsFailure() bool {
	return false
}

// OrElse returns the Try value if success or the value passed as parameter if the Try is a failure.
// for the Success implementation the defined value of the Try is returned.
func (s Success[A]) OrElse(value A) A {
	return s.value
}

// OrElseCause returns the Try cause if failure or the error cause passed as parameter if the Try is a success.
// for the Success implementation the parameter cause is returned.
func (s Success[A]) OrElseCause() (A, error) {
	return s.value, nil
}

// Filter returns an Try containing the value if it matches the predicate or a failure Try.
// for the Success implementation the current option is returned if the value passed the predicate otherwise an failure try is returned.
func (s Success[A]) Filter(predicate func(A) bool, cause error) Try[A] {
	if predicate == nil {
		return FailureOf[A](errors.New("predicate should not be nil"))
	}
	if cause == nil {
		return FailureOf[A](errors.New("error cause should not be nil"))
	}
	if predicate(s.value) {
		return s
	}
	return FailureOf[A](cause)
}

// Failure is an implementation of Try with an error cause.
type Failure[A any] struct {
	cause error
}

// IsFailure checks if the current Try is a failure or not.
// for the Failure implementation the value return is true.
func (f Failure[A]) IsFailure() bool {
	return true
}

// OrElse returns the Try value if success or the value passed as parameter if the Try is a failure.
// for the Failure implementation the parameter value is returned.
func (f Failure[A]) OrElse(value A) A {
	return value
}

// OrElseCause returns the Try cause if failure or the error cause passed as parameter if the Try is a success.
// for the Failure implementation the parameter cause is returned.
func (f Failure[A]) OrElseCause() (A, error) {
	return *new(A), f.cause
}

// Filter returns an Try containing the value if it matches the predicate or a failure Try.
// for the Failure implementation the current failure Try is returned.
func (f Failure[A]) Filter(predicate func(A) bool, cause error) Try[A] {
	if predicate == nil {
		return FailureOf[A](errors.New("predicate should not be nil"))
	}
	if cause == nil {
		return FailureOf[A](errors.New("error cause should not be nil"))
	}
	return f
}

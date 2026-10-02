package control

import (
	"errors"
	"slices"

	"github.com/glours/go2funk/api/tuple"
)

// Validation holds either a valid value or every error found while validating
// it.
//
// Either and Result short-circuit: the first Left wins and every later step is
// skipped. Validation accumulates instead: Check, Sequence and Zip run every
// validation they are given and collect all the errors, which is what a form, a
// config file or a request body needs — reporting one problem at a time is a
// poor experience.
//
// The errors are kept as a slice of E rather than combined through a function
// or a constraint on E. That is what Vavr ends up with too, a Seq of errors, and
// it asks nothing of the error type.
//
// The zero value is valid and holds the zero value of T: no error collected
// means nothing was found wrong. Every operation is O(1) except where its
// documentation says otherwise; those that copy or gather errors are O(k) in the
// number of errors.
type Validation[E, T any] struct {
	value T
	// errors is empty if and only if the validation is valid. It is never
	// mutated once built, which is what lets Map and FlatMap share it.
	errors []E
}

// Valid returns a valid Validation holding value.
func Valid[E, T any](value T) Validation[E, T] {
	return Validation[E, T]{value: value}
}

// Invalid returns an invalid Validation holding err and any further errors, in
// order. Asking for the first error separately is what guarantees there is at
// least one. The errors are copied, so the caller's slice can be reused, into a
// slice allocated once at the right size. O(k).
func Invalid[E, T any](err E, more ...E) Validation[E, T] {
	errs := make([]E, 0, 1+len(more))
	errs = append(errs, err)
	return Validation[E, T]{errors: append(errs, more...)}
}

// IsValid reports whether the validation holds a value.
func (v Validation[E, T]) IsValid() bool { return len(v.errors) == 0 }

// IsInvalid reports whether the validation holds errors.
func (v Validation[E, T]) IsInvalid() bool { return len(v.errors) > 0 }

// Get returns the value and whether the validation is valid. An invalid
// validation yields the zero value of T.
func (v Validation[E, T]) Get() (T, bool) {
	if v.IsInvalid() {
		var zero T
		return zero, false
	}
	return v.value, true
}

// Errors returns a copy of the errors, in the order they were collected, or
// nil when the validation is valid. O(k).
func (v Validation[E, T]) Errors() []E { return slices.Clone(v.errors) }

// OrElse returns the value if the validation is valid, otherwise other.
func (v Validation[E, T]) OrElse(other T) T {
	if v.IsInvalid() {
		return other
	}
	return v.value
}

// Map applies mapper to the value. An invalid validation keeps its errors and
// mapper is not called.
func (v Validation[E, T]) Map[U any](mapper func(T) U) Validation[E, U] {
	if v.IsInvalid() {
		return Validation[E, U]{errors: v.errors}
	}
	return Valid[E](mapper(v.value))
}

// MapError applies mapper to every error, keeping their order. A valid
// validation is returned unchanged and mapper is not called. O(k).
func (v Validation[E, T]) MapError[F any](mapper func(E) F) Validation[F, T] {
	if v.IsValid() {
		return Valid[F](v.value)
	}
	mapped := make([]F, len(v.errors))
	for index, err := range v.errors {
		mapped[index] = mapper(err)
	}
	return Validation[F, T]{errors: mapped}
}

// FlatMap applies mapper to the value and returns its result. It is sequential
// by nature — the next step needs the value — so an invalid validation is
// returned with its own errors and mapper is not called. To run independent
// validations and collect all their errors, use Check, Sequence or Zip.
func (v Validation[E, T]) FlatMap[U any](mapper func(T) Validation[E, U]) Validation[E, U] {
	if v.IsInvalid() {
		return Validation[E, U]{errors: v.errors}
	}
	return mapper(v.value)
}

// Fold returns onValid applied to the value, or onInvalid applied to a copy of
// the errors. Exactly one of the two functions is called. O(k) when invalid.
func (v Validation[E, T]) Fold[U any](onInvalid func([]E) U, onValid func(T) U) U {
	if v.IsInvalid() {
		return onInvalid(v.Errors())
	}
	return onValid(v.value)
}

// Check runs every check against value and collects all their errors, in the
// order of the checks. It returns value, valid, when no check found anything
// wrong. The checks validate, they do not transform: what a check returns on
// success is ignored, so a normalising step belongs in Map or FlatMap.
// O(c + k) for c checks and k errors.
func Check[E, T any](value T, checks ...func(T) Validation[E, T]) Validation[E, T] {
	// errs belongs to this call alone, so appending to it cannot reach into
	// a validation a check returned.
	var errs []E
	for _, check := range checks {
		errs = append(errs, check(value).errors...)
	}
	return Validation[E, T]{value: value, errors: errs}
}

// Sequence turns validations of the same type into a validation of all their
// values, in order, or of all their errors, in order. Every validation is
// looked at, so nothing is lost to an early failure.
//
// A first pass counts the errors, so Sequence knows before building anything
// whether it gathers values or errors, and how many: a single allocation of the
// exact size either way. O(n + k).
func Sequence[E, T any](validations ...Validation[E, T]) Validation[E, []T] {
	count := 0
	for _, validation := range validations {
		count += len(validation.errors)
	}

	if count > 0 {
		errs := make([]E, 0, count)
		for _, validation := range validations {
			errs = append(errs, validation.errors...)
		}
		return Validation[E, []T]{errors: errs}
	}

	values := make([]T, len(validations))
	for index, validation := range validations {
		values[index] = validation.value
	}
	return Valid[E](values)
}

// Zip pairs two validations of different types: both values when both are
// valid, otherwise the errors of both, v's first. Zipping the result again
// combines any number of validations without an arity limit, at the cost of
// nesting the pairs. O(k).
//
// It is a function rather than a method because a method cannot do this in Go:
// a Zip on Validation[E, T] returning Validation[E, Pair[T, U]] would give that
// type a Zip of its own, returning Validation[E, Pair[Pair[T, U], V]], and so
// on without end — the compiler rejects the instantiation cycle.
func Zip[E, T, U any](v Validation[E, T], other Validation[E, U]) Validation[E, tuple.Pair[T, U]] {
	if v.IsInvalid() || other.IsInvalid() {
		// A fresh slice: appending to v.errors would write into its spare
		// capacity, which another Zip from the same v could be using too.
		return Validation[E, tuple.Pair[T, U]]{errors: slices.Concat(v.errors, other.errors)}
	}
	return Valid[E](tuple.New(v.value, other.value))
}

// FromEither converts an Either: a Right becomes valid, a Left becomes invalid
// with its value as the only error. Result is an Either, so this converts a
// Result too, keeping its cause.
func FromEither[E, T any](either Either[E, T]) Validation[E, T] {
	return either.Fold(
		func(err E) Validation[E, T] { return Invalid[E, T](err) },
		Valid[E, T],
	)
}

// ToEither converts to an Either holding the value on the right, or a copy of
// every error on the left. O(k).
func (v Validation[E, T]) ToEither() Either[[]E, T] {
	if v.IsInvalid() {
		return Left[[]E, T](v.Errors())
	}
	return Right[[]E](v.value)
}

// ToOption returns Some of the value, or None when invalid. The errors are
// discarded.
func (v Validation[E, T]) ToOption() Option[T] {
	if v.IsInvalid() {
		return None[T]()
	}
	return Some(v.value)
}

// ToResult converts a validation of errors to a Result, joining every error
// with errors.Join so that errors.Is and errors.As still reach each cause.
// Unwrap then brings it back to the (T, error) idiom. O(k).
//
// It is a function because a method cannot require E to be error. errors.Join
// drops nil errors, so an invalid validation holding only nil ones gives a
// failed Result whose cause is nil, as Err(nil) does.
func ToResult[T any](v Validation[error, T]) Result[T] {
	if v.IsInvalid() {
		return Err[T](errors.Join(v.errors...))
	}
	return Ok(v.value)
}

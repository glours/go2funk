// Package control provides control structures such as Option, Either and Result.
//
// Every type in this package is a concrete struct, not an interface: Go 1.27
// allows methods to declare their own type parameters, but only on concrete
// types, which is what makes Map able to change the type it carries.
//
// A useful consequence is that the zero value of every type is meaningful and
// safe to use. All operations are O(1).
package control

// Option represents an optional value: either Some, holding a value, or None.
//
// The zero value is None, so an Option field needs no initialisation:
//
//	type Config struct{ Timeout Option[time.Duration] }
//	var c Config
//	c.Timeout.OrElse(30 * time.Second) // works, no panic
type Option[T any] struct {
	value   T
	defined bool
}

// Some returns an Option holding value.
func Some[T any](value T) Option[T] {
	return Option[T]{value: value, defined: true}
}

// None returns an empty Option. It is the same as the zero value of Option[T].
func None[T any]() Option[T] {
	return Option[T]{}
}

// FromPointer returns Some(*pointer), or None if pointer is nil.
func FromPointer[T any](pointer *T) Option[T] {
	if pointer == nil {
		return None[T]()
	}
	return Some(*pointer)
}

// FromTuple lifts the Go "comma ok" idiom into an Option.
//
//	value, ok := myMap[key]
//	opt := FromTuple(value, ok)
func FromTuple[T any](value T, ok bool) Option[T] {
	if !ok {
		return None[T]()
	}
	return Some(value)
}

// IsDefined reports whether a value is present.
func (o Option[T]) IsDefined() bool { return o.defined }

// IsEmpty reports whether no value is present.
func (o Option[T]) IsEmpty() bool { return !o.defined }

// Get returns the value and whether it was present, following the Go
// "comma ok" idiom. The value is the zero value of T when absent.
func (o Option[T]) Get() (T, bool) { return o.value, o.defined }

// OrElse returns the value if present, otherwise other.
func (o Option[T]) OrElse(other T) T {
	if o.defined {
		return o.value
	}
	return other
}

// OrElseGet returns the value if present, otherwise the result of other.
// other is only evaluated when the Option is empty.
func (o Option[T]) OrElseGet(other func() T) T {
	if o.defined {
		return o.value
	}
	return other()
}

// OrElseError returns the value and a nil error if present, otherwise the zero
// value of T and err. It is the bridge back to the Go error idiom.
func (o Option[T]) OrElseError(err error) (T, error) {
	if o.defined {
		return o.value, nil
	}
	var zero T
	return zero, err
}

// Or returns the current Option if a value is present, otherwise other.
func (o Option[T]) Or(other Option[T]) Option[T] {
	if o.defined {
		return o
	}
	return other
}

// Filter returns the current Option if its value matches the predicate,
// otherwise an empty Option. The predicate is not called on an empty Option.
func (o Option[T]) Filter(predicate func(T) bool) Option[T] {
	if o.defined && predicate(o.value) {
		return o
	}
	return None[T]()
}

// ForEach calls action on the value, if there is one.
func (o Option[T]) ForEach(action func(T)) {
	if o.defined {
		action(o.value)
	}
}

// ToPointer returns a pointer to a copy of the value, or nil if empty.
func (o Option[T]) ToPointer() *T {
	if !o.defined {
		return nil
	}
	value := o.value
	return &value
}

// ToSlice returns a slice holding the value, or an empty slice.
func (o Option[T]) ToSlice() []T {
	if !o.defined {
		return []T{}
	}
	return []T{o.value}
}

// Map applies mapper to the value and returns an Option of the result, or an
// empty Option[U] if there is no value. mapper is not called on an empty Option.
func (o Option[T]) Map[U any](mapper func(T) U) Option[U] {
	if !o.defined {
		return None[U]()
	}
	return Some(mapper(o.value))
}

// FlatMap applies mapper to the value and returns its result, or an empty
// Option[U] if there is no value. mapper is not called on an empty Option.
func (o Option[T]) FlatMap[U any](mapper func(T) Option[U]) Option[U] {
	if !o.defined {
		return None[U]()
	}
	return mapper(o.value)
}

// Fold returns onSome applied to the value, or onNone() if there is no value.
// Exactly one of the two functions is called.
func (o Option[T]) Fold[U any](onNone func() U, onSome func(T) U) U {
	if !o.defined {
		return onNone()
	}
	return onSome(o.value)
}

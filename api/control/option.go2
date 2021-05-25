// Package control provides control structures such as Option, Try or Either...
package control

// Option is a container interface which represents a optional value.
// internal implementations of Option are Some and None.
type Option[T any] interface {
	IsEmpty() bool
	OrElse(value T) T
	OrElseError(err error) (T, error)
	Filter(func(T) bool) Option[T]
}

// MapOption maps the element of an Option[T] to a new option with element of type U.
// the mapper function should take a T value and return a U value.
func MapOption[T, U any](option Option[T], mapper func(T) U) Option[U] {
	if option.IsEmpty() {
		return None[U]{}
	}
	return Some[U]{mapper(option.OrElse(*new(T)))}
}

// FlatMapOption maps the element of an Option[T] to a new option with element of type U.
// the mapper function should take a T value and return an Option[U] as result.
func FlatMapOption[T, U any](option Option[T], mapper func(T) Option[U]) Option[U] {
	if option.IsEmpty() {
		return Empty[U]()
	}
	return mapper(option.OrElse(*new(T)))
}

// Empty returns a None[T] as Option[T].
func Empty[T any]() Option[T] {
	return None[T]{}
}

// Of returns a Some[T] with the parameter value as Option[T].
func Of[T any](value T) Option[T] {
	return Some[T]{value}
}

// None is an implementation of a Option with an undefined value.
type None[T any] struct{}

// IsEmpty checks if the current Option is empty.
// for the None implementation the value returned is true.
func (n None[T]) IsEmpty() bool {
	return true
}

// OrElse returns the Option value if defined or the value passed as parameter if the Option is empty.
// for the None implementation the parameter value is returned.
func (n None[T]) OrElse(value T) T {
	return value
}

// OrElseError returns the Option value if defined or the error passed as parameter if the Option is empty.
// for the None implementation the error parameter is returned.
func (n None[T]) OrElseError(err error) (T, error) {
	return *new(T), err
}

// Filter returns an Option containing the value if it matches the predicate or an empty Option.
// for the None implementation an new empty Option is returned.
func (n None[T]) Filter(predictate func(T) bool) Option[T] {
	return Empty[T]()
}

// Some is an implementation of a Option with a defined value.
type Some[T any] struct {
	value T
}

// IsEmpty checks if the current Option is empty.
// for the Some implementation the value returned is false.
func (s Some[T]) IsEmpty() bool {
	return false
}

// OrElse returns the Option value if defined or the value passed as parameter if the Option is empty.
// for the Some implementation the value of the current Option is returned.
func (s Some[T]) OrElse(value T) T {
	return s.value
}

// OrElseError returns the Option value if defined or the error passed as parameter if the Option is empty.
// for the Some implementation the option value is returned.
func (s Some[T]) OrElseError(err error) (T, error) {
	return s.value, nil
}

// Filter returns an Option containing the value if it matches the predicate or an empty Option.
// for the Some implementation the current option is returned if the value passed the predicate otherwise an empty option is returned.
func (s Some[T]) Filter(predicate func(T) bool) Option[T] {
	if predicate(s.value) {
		return s
	}
	return Empty[T]()
}

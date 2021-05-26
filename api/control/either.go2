// Package control provides control structures such as Option, Try or Either...
package control

// Either represents a value of two possible types.
// internal implementations of Either are Right and Left.
type Either[L, R any] interface {
	IsLeft() bool
	IsRight() bool
	OrElse(Either[L, R]) Either[L, R]
	GetOrElse(other R) R
	GetLeftOrElse(other L) L
	Swap() Either[R, L]
	FilterOrElse(func(R) bool, func(R) L) Either[L, R]
	Filter(func(R) bool) Option[Either[L, R]]
}

// MapEither maps the Right element of a Either[L,R] to a new Either with a right element of type U.
// the mapper function should take a R value and return a U value.
func MapEither[L, R, U any](either Either[L, R], mapper func(R) U) Either[L, U] {
	if either.IsRight() {
		return Right[L, U]{mapper(either.GetOrElse(*new(R)))}
	}
	return Left[L, U]{either.GetLeftOrElse(*new(L))}
}

// FlatMapEither maps the Right element of a Either[L,R] to a new Either with a right element of type U.
// the mapper function should take a R value and return a Either[L,U] value.
func FlatMapEither[L, R, U any](either Either[L, R], mapper func(R) Either[L, U]) Either[L, U] {
	if either.IsRight() {
		return mapper(either.GetOrElse(*new(R)))
	}
	return Left[L, U]{either.GetLeftOrElse(*new(L))}
}

// RightOf return a Either[L,R] with the right value set.
func RightOf[L, R any](value R) Either[L, R] {
	return Right[L, R]{value}
}

// LeftOf return a Either[L,R] with the left value set.
func LeftOf[L, R any](value L) Either[L, R] {
	return Left[L, R]{value}
}

// Right is an implementation of Either with a "right" value initialized
type Right[L, R any] struct {
	value R
}

// IsLeft checks if the current Either contains a "left" value or not
// Right implementation always return false
func (r Right[L, R]) IsLeft() bool {
	return false
}

// IsRight checks if the current Either contains a "right" value or not
// Right implementation always return true
func (r Right[L, R]) IsRight() bool {
	return true
}

// Swap converts a Right Either to a Left one and vis versa
// Right implementation return a new Left Either setup with the previous "right" value.
func (r Right[L, R]) Swap() Either[R, L] {
	return LeftOf[R, L](r.value)
}

// OrElse returns the current Either if it's a Right one or the Either passed as parameter.
// Right implementation always return the current Either.
func (r Right[L, R]) OrElse(other Either[L, R]) Either[L, R] {
	return r
}

// GetOrElse returns the "right" value of the current Either or the "other" value passed as parameter.
// Right implementation always return the current Either "right" value.
func (r Right[L, R]) GetOrElse(other R) R {
	return r.value
}

// FilterOrElse turns a Right Either into a Left one if the "right" value does not make it through the predicate.
func (r Right[L, R]) FilterOrElse(predicate func(R) bool, transform func(R) L) Either[L, R] {
	if predicate(r.value) {
		return r
	}
	return LeftOf[L, R](transform(r.value))
}

// Filter returns an Option with the current either if the "right" value matches the predicate.
func (r Right[L, R]) Filter(predicate func(R) bool) Option[Either[L, R]] {
	if predicate(r.value) {
		return Of[Either[L, R]](r)
	}
	return Empty[Either[L, R]]()
}

// GetLeftOrElse return the "left" value of a Left Either or the "other" value passed as parameter.
// Right implementation always return the "other" value passed as parameter.
func (r Right[L, R]) GetLeftOrElse(other L) L {
	return other
}

// Left is an implementation of Either with a "left" value initialized
type Left[L, R any] struct {
	value L
}

// IsLeft checks if the current Either contains a "left" value or not
// Left implementation always return true
func (l Left[L, R]) IsLeft() bool {
	return true
}

// IsRight checks if the current Either contains a "right" value or not
// Left implementation always return false
func (l Left[L, R]) IsRight() bool {
	return false
}

// Swap converts a right Either to a left one and vis versa
// Left implementation return a new Right Either setup with the previous "left" value.
func (l Left[L, R]) Swap() Either[R, L] {
	return RightOf[R, L](l.value)
}

// OrElse returns the current Either if it's a Right one or the Either passed as parameter.
// Left implementation always return the Either passed as parameter.
func (l Left[L, R]) OrElse(other Either[L, R]) Either[L, R] {
	return other
}

// GetOrElse returns the "right" value of the current Either or the "other" value passed as parameter.
// Left implementation always return the "other" value passed as parameter.
func (l Left[L, R]) GetOrElse(other R) R {
	return other
}

// FilterOrElse turns a Right Either into a Left one if the "right" value does not make it through the predicate.
// Left implementation always return the current Either.
func (l Left[L, R]) FilterOrElse(predicate func(R) bool, transform func(R) L) Either[L, R] {
	return l
}

// Filter returns an Option with the current either if the "right" value matches the predicate.
// Left implementation always returns a empty Option.
func (l Left[L, R]) Filter(predicate func(R) bool) Option[Either[L, R]] {
	return Empty[Either[L, R]]()
}

// GetLeftOrElse return the "left" value of a Left Either or the "other" value passed as parameter.
// Left implementation always return the left value of the current Either.
func (l Left[L, R]) GetLeftOrElse(other L) L {
	return l.value
}

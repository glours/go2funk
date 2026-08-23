// Package tuple provides product types: values that hold several others at once.
//
// All operations are O(1) and return new values; nothing is mutated in place.
package tuple

// Pair holds two values of independent types.
//
// The zero value holds the zero values of both sides and is safe to use.
type Pair[L, R any] struct {
	left  L
	right R
}

// New returns a Pair holding left and right.
func New[L, R any](left L, right R) Pair[L, R] {
	return Pair[L, R]{left: left, right: right}
}

// Left returns the left value.
func (p Pair[L, R]) Left() L { return p.left }

// Right returns the right value.
func (p Pair[L, R]) Right() R { return p.right }

// Unpack returns both values at once, for destructuring assignment.
func (p Pair[L, R]) Unpack() (L, R) { return p.left, p.right }

// Swap returns a Pair with the two sides exchanged.
func (p Pair[L, R]) Swap() Pair[R, L] {
	return Pair[R, L]{left: p.right, right: p.left}
}

// MapLeft applies mapper to the left value and leaves the right one untouched.
func (p Pair[L, R]) MapLeft[U any](mapper func(L) U) Pair[U, R] {
	return Pair[U, R]{left: mapper(p.left), right: p.right}
}

// MapRight applies mapper to the right value and leaves the left one untouched.
func (p Pair[L, R]) MapRight[U any](mapper func(R) U) Pair[L, U] {
	return Pair[L, U]{left: p.left, right: mapper(p.right)}
}

// Map applies one mapper to each side.
func (p Pair[L, R]) Map[T, U any](mapLeft func(L) T, mapRight func(R) U) Pair[T, U] {
	return Pair[T, U]{left: mapLeft(p.left), right: mapRight(p.right)}
}

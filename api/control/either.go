package control

// Either holds one of two values. By convention Right carries the expected
// value and Left the alternative one, so Map, FlatMap and Filter all operate on
// the Right side and let a Left through untouched.
//
// The zero value is a Left holding the zero value of L.
type Either[L, R any] struct {
	left    L
	right   R
	isRight bool
}

// Right returns an Either holding value on its right side.
func Right[L, R any](value R) Either[L, R] {
	return Either[L, R]{right: value, isRight: true}
}

// Left returns an Either holding value on its left side.
func Left[L, R any](value L) Either[L, R] {
	return Either[L, R]{left: value}
}

// IsRight reports whether the right side is set.
func (e Either[L, R]) IsRight() bool { return e.isRight }

// IsLeft reports whether the left side is set.
func (e Either[L, R]) IsLeft() bool { return !e.isRight }

// Get returns the right value and whether it was set.
func (e Either[L, R]) Get() (R, bool) { return e.right, e.isRight }

// GetLeft returns the left value and whether it was set.
func (e Either[L, R]) GetLeft() (L, bool) { return e.left, !e.isRight }

// OrElse returns the right value if set, otherwise other.
func (e Either[L, R]) OrElse(other R) R {
	if e.isRight {
		return e.right
	}
	return other
}

// OrElseGet returns the right value if set, otherwise the result of other
// applied to the left value. other is only evaluated on a Left.
func (e Either[L, R]) OrElseGet(other func(L) R) R {
	if e.isRight {
		return e.right
	}
	return other(e.left)
}

// LeftOrElse returns the left value if set, otherwise other.
func (e Either[L, R]) LeftOrElse(other L) L {
	if e.isRight {
		return other
	}
	return e.left
}

// Or returns the current Either if it is a Right, otherwise other.
func (e Either[L, R]) Or(other Either[L, R]) Either[L, R] {
	if e.isRight {
		return e
	}
	return other
}

// Swap exchanges the two sides.
func (e Either[L, R]) Swap() Either[R, L] {
	if e.isRight {
		return Left[R, L](e.right)
	}
	return Right[R, L](e.left)
}

// FilterOrElse turns a Right into a Left, built by orElse, when its value does
// not match the predicate. A Left is returned unchanged and neither function is
// called.
func (e Either[L, R]) FilterOrElse(predicate func(R) bool, orElse func(R) L) Either[L, R] {
	if !e.isRight || predicate(e.right) {
		return e
	}
	return Left[L, R](orElse(e.right))
}

// ForEach calls action on the right value, if there is one.
func (e Either[L, R]) ForEach(action func(R)) {
	if e.isRight {
		action(e.right)
	}
}

// ToOption returns Some of the right value, or None on a Left. The left value
// is discarded.
func (e Either[L, R]) ToOption() Option[R] {
	if !e.isRight {
		return None[R]()
	}
	return Some(e.right)
}

// Map applies mapper to the right value. A Left is returned unchanged, keeping
// its value, and mapper is not called.
func (e Either[L, R]) Map[U any](mapper func(R) U) Either[L, U] {
	if !e.isRight {
		return Left[L, U](e.left)
	}
	return Right[L, U](mapper(e.right))
}

// FlatMap applies mapper to the right value and returns its result. A Left is
// returned unchanged and mapper is not called.
func (e Either[L, R]) FlatMap[U any](mapper func(R) Either[L, U]) Either[L, U] {
	if !e.isRight {
		return Left[L, U](e.left)
	}
	return mapper(e.right)
}

// MapLeft applies mapper to the left value. A Right is returned unchanged.
func (e Either[L, R]) MapLeft[U any](mapper func(L) U) Either[U, R] {
	if e.isRight {
		return Right[U, R](e.right)
	}
	return Left[U, R](mapper(e.left))
}

// Fold returns onRight applied to the right value, or onLeft applied to the
// left one. Exactly one of the two functions is called.
func (e Either[L, R]) Fold[U any](onLeft func(L) U, onRight func(R) U) U {
	if !e.isRight {
		return onLeft(e.left)
	}
	return onRight(e.right)
}

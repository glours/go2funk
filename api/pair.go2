package api

type Pair[L, R any] struct {
	left  L
	right R
}

func NewPair[L, R any](left L, right R) Pair[L, R] {
	return Pair[L, R]{left, right}
}

func MapPair[L, R, T, U any](pair Pair[L, R], mapperLeft func(L) T, mapperRight func(R) U) Pair[T, U] {
	return Pair[T, U]{ mapperLeft(pair.left), mapperRight(pair.right)}
}

func MapLeftPair[L, R, U any](pair Pair[L, R], mapper func(L) U) Pair[U, R] {
	return Pair[U, R]{mapper(pair.left), pair.right}
}

func MapRightPair[L, R, U any](pair Pair[L, R], mapper func(R) U) Pair[L, U] {
	return Pair[L, U]{pair.left, mapper(pair.right)}
}

func (p Pair[L, R]) GetLeft() L {
	return p.left
}

func (p Pair[L, R]) GetRight() R {
	return p.right
}
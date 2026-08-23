package control

// Result is an Either whose left side is an error. It is a type alias, not a
// distinct type, so a Result[T] is an Either[error, T] and inherits every one of
// its methods — Map, FlatMap, Fold, ToOption and the rest — including the
// ability to change the type it carries.
//
// The trade-off of the alias is that Result cannot declare methods of its own;
// the helpers below are package-level functions for that reason.
type Result[T any] = Either[error, T]

// Ok returns a successful Result holding value.
func Ok[T any](value T) Result[T] {
	return Right[error, T](value)
}

// Err returns a failed Result holding cause.
func Err[T any](cause error) Result[T] {
	return Left[error, T](cause)
}

// Try runs f and captures its outcome: Ok on success, Err holding the returned
// error on failure. The error is never discarded.
func Try[T any](f func() (T, error)) Result[T] {
	value, err := f()
	if err != nil {
		return Err[T](err)
	}
	return Ok(value)
}

// Unwrap converts a Result back to the Go (value, error) idiom. A failed Result
// yields the zero value of T and its cause.
func Unwrap[T any](result Result[T]) (T, error) {
	if value, ok := result.Get(); ok {
		return value, nil
	}
	cause, _ := result.GetLeft()
	var zero T
	return zero, cause
}

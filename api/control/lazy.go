package control

import (
	"sync"
	"sync/atomic"
)

// Lazy defers a computation until its result is first read, then remembers it.
// The computation runs at most once, however many goroutines ask for it.
//
//	config := control.NewLazy(loadConfig) // nothing has run yet
//	config.Get()                          // runs loadConfig
//	config.Get()                          // returns the remembered result
//
// A Lazy is a value: copying one shares the memoised result rather than
// restarting the computation.
//
// The zero value has no computation attached and yields the zero value of T,
// the way Rust's LazyCell::default() does. It never panics. Lazy carries no
// notion of absence: when "not computed yet" has to be told apart from
// "computed to the zero value", the type to reach for is Lazy[Option[T]].
// Likewise a computation that can fail is a Lazy[Result[T]].
type Lazy[T any] struct {
	state *lazyState[T]
}

// lazyState is shared by every copy of a Lazy. It lives behind a pointer so
// that Lazy stays a plain value: the sync.Once is never copied.
type lazyState[T any] struct {
	once      sync.Once
	compute   func() T
	value     T
	evaluated atomic.Bool
}

// NewLazy returns a Lazy that will call compute on first read.
// A nil compute yields the zero value of T, like the zero value of Lazy.
func NewLazy[T any](compute func() T) Lazy[T] {
	if compute == nil {
		return Lazy[T]{}
	}
	return Lazy[T]{state: &lazyState[T]{compute: compute}}
}

// Delay is an alias for NewLazy, under the name functional languages give it.
func Delay[T any](compute func() T) Lazy[T] {
	return NewLazy(compute)
}

// Get returns the result of the computation, running it on the first call.
// It is safe to call from several goroutines: the computation runs once and
// every caller sees the same result.
func (l Lazy[T]) Get() T {
	if l.state == nil {
		var zero T
		return zero
	}
	l.state.once.Do(func() {
		l.state.value = l.state.compute()
		l.state.compute = nil // let the closure and what it captured go
		l.state.evaluated.Store(true)
	})
	return l.state.value
}

// IsEvaluated reports whether the result is already available. A Lazy with no
// computation attached has nothing to evaluate, so it reports true.
func (l Lazy[T]) IsEvaluated() bool {
	if l.state == nil {
		return true
	}
	return l.state.evaluated.Load()
}

// Map returns a Lazy of mapper applied to the result. Neither the receiver nor
// mapper runs until the returned Lazy is read.
func (l Lazy[T]) Map[U any](mapper func(T) U) Lazy[U] {
	return NewLazy(func() U { return mapper(l.Get()) })
}

// FlatMap returns the Lazy produced by mapper. Nothing runs until the returned
// Lazy is read.
func (l Lazy[T]) FlatMap[U any](mapper func(T) Lazy[U]) Lazy[U] {
	return NewLazy(func() U { return mapper(l.Get()).Get() })
}

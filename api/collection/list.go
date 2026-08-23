// Package collection provides purely functional collections.
//
// The collections here are persistent: every operation returns a new value and
// nothing is ever mutated in place, so sharing a collection between goroutines
// or keeping an old version around is always safe.
//
// Every type integrates with the Go iterator protocol through All, which
// returns an iter.Seq usable directly in a for range loop.
package collection

import (
	"fmt"
	"iter"
	"strings"

	"github.com/glours/go2funk/api/control"
)

// node is one cell of the singly linked list. size is the number of elements
// from this cell to the end, which makes Length O(1).
type node[T any] struct {
	value T
	next  *node[T]
	size  int
}

// List is an immutable, persistent singly linked list.
//
// The zero value is the empty list, so a List field needs no initialisation.
// Prepend, Head, Tail, Length and IsEmpty are O(1); everything that has to walk
// the list is O(n) and says so.
type List[T any] struct {
	head *node[T]
}

// Empty returns the empty list. It is the same as the zero value of List[T].
func Empty[T any]() List[T] {
	return List[T]{}
}

// Of returns a list holding values, in order. Called with no argument it
// returns the empty list. O(n).
func Of[T any](values ...T) List[T] {
	list := Empty[T]()
	for i := len(values) - 1; i >= 0; i-- {
		list = list.Prepend(values[i])
	}
	return list
}

// Collect drains seq into a new list, preserving order. O(n).
func Collect[T any](seq iter.Seq[T]) List[T] {
	var reversed List[T]
	for value := range seq {
		reversed = reversed.Prepend(value)
	}
	return reversed.Reverse()
}

// Remove returns a list without any element equal to value. The original list
// is left untouched. O(n).
//
// It is a function rather than a method because it needs T to be comparable,
// which a method of List[T any] cannot require.
func Remove[T comparable](list List[T], value T) List[T] {
	return list.Filter(func(candidate T) bool { return candidate != value })
}

// IsEmpty reports whether the list holds no element. O(1).
func (l List[T]) IsEmpty() bool { return l.head == nil }

// Length returns the number of elements. O(1).
func (l List[T]) Length() int {
	if l.head == nil {
		return 0
	}
	return l.head.size
}

// Head returns the first element, or None if the list is empty. O(1).
func (l List[T]) Head() control.Option[T] {
	if l.head == nil {
		return control.None[T]()
	}
	return control.Some(l.head.value)
}

// Tail returns the list without its first element. The tail of the empty list
// is the empty list. O(1).
func (l List[T]) Tail() List[T] {
	if l.head == nil {
		return l
	}
	return List[T]{head: l.head.next}
}

// Get returns the element at index, or None if the index is out of range. O(n).
func (l List[T]) Get(index int) control.Option[T] {
	if index < 0 {
		return control.None[T]()
	}
	for current := l.head; current != nil; current = current.next {
		if index == 0 {
			return control.Some(current.value)
		}
		index--
	}
	return control.None[T]()
}

// Prepend returns a new list with value in front. O(1).
func (l List[T]) Prepend(value T) List[T] {
	size := 1
	if l.head != nil {
		size += l.head.size
	}
	return List[T]{head: &node[T]{value: value, next: l.head, size: size}}
}

// Append returns a new list with value at the end. O(n): a singly linked list
// has to be rebuilt to grow at the tail. Prefer Prepend when order allows it.
func (l List[T]) Append(value T) List[T] {
	return l.AppendAll(value)
}

// AppendAll returns a new list with values added at the end, in order. O(n+m).
func (l List[T]) AppendAll(values ...T) List[T] {
	result := l.reversed()
	for _, value := range values {
		result = result.Prepend(value)
	}
	return result.Reverse()
}

// Reverse returns the list in reverse order. O(n).
func (l List[T]) Reverse() List[T] { return l.reversed() }

// reversed builds the reversed list by prepending, which is O(n) overall.
func (l List[T]) reversed() List[T] {
	var result List[T]
	for current := l.head; current != nil; current = current.next {
		result = result.Prepend(current.value)
	}
	return result
}

// Filter returns a list of the elements matching the predicate, in order. O(n).
func (l List[T]) Filter(predicate func(T) bool) List[T] {
	var result List[T]
	for current := l.head; current != nil; current = current.next {
		if predicate(current.value) {
			result = result.Prepend(current.value)
		}
	}
	return result.Reverse()
}

// Insert returns a new list with value at index. It reports an error when index
// is negative or greater than the length. O(n).
func (l List[T]) Insert(index int, value T) (List[T], error) {
	if index < 0 || index > l.Length() {
		return l, fmt.Errorf("index out of range %d on a List of length %d", index, l.Length())
	}

	var head List[T]
	current := l.head
	for i := 0; i < index; i++ {
		head = head.Prepend(current.value)
		current = current.next
	}

	result := List[T]{head: current}.Prepend(value)
	for node := head.head; node != nil; node = node.next {
		result = result.Prepend(node.value)
	}
	return result, nil
}

// All returns an iterator over the elements, in order, for use with for range.
func (l List[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for current := l.head; current != nil; current = current.next {
			if !yield(current.value) {
				return
			}
		}
	}
}

// ToSlice returns the elements as a new slice. O(n).
func (l List[T]) ToSlice() []T {
	result := make([]T, 0, l.Length())
	for current := l.head; current != nil; current = current.next {
		result = append(result, current.value)
	}
	return result
}

// ForEach calls action on every element, in order. O(n).
func (l List[T]) ForEach(action func(T)) {
	for current := l.head; current != nil; current = current.next {
		action(current.value)
	}
}

// String renders the list as List(a, b, c).
func (l List[T]) String() string {
	var b strings.Builder
	b.WriteString("List(")
	for current, first := l.head, true; current != nil; current, first = current.next, false {
		if !first {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%v", current.value)
	}
	b.WriteString(")")
	return b.String()
}

// Map returns the list of mapper applied to every element, in order. O(n).
func (l List[T]) Map[U any](mapper func(T) U) List[U] {
	var result List[U]
	for current := l.head; current != nil; current = current.next {
		result = result.Prepend(mapper(current.value))
	}
	return result.Reverse()
}

// FlatMap applies mapper to every element and concatenates the results. O(n+m).
func (l List[T]) FlatMap[U any](mapper func(T) List[U]) List[U] {
	var result List[U]
	for current := l.head; current != nil; current = current.next {
		for value := range mapper(current.value).All() {
			result = result.Prepend(value)
		}
	}
	return result.Reverse()
}

// Fold combines the elements from the left, starting from initial. O(n).
func (l List[T]) Fold[U any](initial U, combine func(U, T) U) U {
	result := initial
	for current := l.head; current != nil; current = current.next {
		result = combine(result, current.value)
	}
	return result
}

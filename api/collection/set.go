package collection

import (
	"fmt"
	"iter"
	"slices"
	"strings"
)

// Set is an immutable, persistent hash set.
//
// It is a Map whose values carry nothing, so it shares the map's trie and its
// guarantees rather than duplicating them. Elements only need to be comparable:
// structs, arrays and pointers all work. Tree is the set to reach for when the
// elements are ordered and iteration should follow that order.
//
// Iteration order is unspecified, as for Map. Equality follows ==, so like a Go
// map key a float NaN can be inserted but never found again.
//
// The zero value is the empty set. Len and IsEmpty are O(1); Contains, Insert and
// Delete are O(log32 n). Union, Intersection and Difference walk the smaller of
// the two sets and look each of its elements up in the larger one, so they cost
// O(m log32 n) for m the smaller size and n the larger: combining a handful of
// elements with a large set stays cheap whichever side it is on.
type Set[T comparable] struct {
	entries Map[T, struct{}]
}

// EmptySet returns the empty set. It is the same as the zero value of Set.
func EmptySet[T comparable]() Set[T] {
	return Set[T]{}
}

// SetOf returns a set holding values. Duplicates collapse into one.
// O(n log32 n).
func SetOf[T comparable](values ...T) Set[T] {
	result := EmptySet[T]()
	for _, value := range values {
		result = result.Insert(value)
	}
	return result
}

// CollectSet drains seq into a new set. O(n log32 n).
func CollectSet[T comparable](seq iter.Seq[T]) Set[T] {
	result := EmptySet[T]()
	for value := range seq {
		result = result.Insert(value)
	}
	return result
}

// IsEmpty reports whether the set holds no element. O(1).
func (s Set[T]) IsEmpty() bool { return s.entries.IsEmpty() }

// Len returns the number of elements. O(1).
func (s Set[T]) Len() int { return s.entries.Len() }

// Contains reports whether value is in the set. O(log32 n).
func (s Set[T]) Contains(value T) bool { return s.entries.ContainsKey(value) }

// Insert returns a set holding value. Inserting a value that is already there
// returns the very same set, sharing every node, as Tree does. Map.Put cannot
// promise that, since it cannot compare values, but a set has no value to
// replace. O(log32 n).
func (s Set[T]) Insert(value T) Set[T] {
	if s.Contains(value) {
		return s
	}
	return Set[T]{entries: s.entries.Put(value, struct{}{})}
}

// Delete returns a set without value. Deleting a value that is not there returns
// the very same set, sharing every node. O(log32 n).
func (s Set[T]) Delete(value T) Set[T] {
	return Set[T]{entries: s.entries.Delete(value)}
}

// Union returns the set of elements in either set. The smaller set is inserted
// into the larger one, which is shared rather than rebuilt.
// O(m log32 n), m being the smaller size.
func (s Set[T]) Union(other Set[T]) Set[T] {
	larger, smaller := s, other
	if smaller.Len() > larger.Len() {
		larger, smaller = smaller, larger
	}
	result := larger
	for value := range smaller.All() {
		result = result.Insert(value)
	}
	return result
}

// Intersection returns the set of elements in both sets. Only the smaller set is
// walked, since no element outside it can be in the result.
// O(m log32 n), m being the smaller size.
func (s Set[T]) Intersection(other Set[T]) Set[T] {
	larger, smaller := s, other
	if smaller.Len() > larger.Len() {
		larger, smaller = smaller, larger
	}
	result := EmptySet[T]()
	for value := range smaller.All() {
		if larger.Contains(value) {
			result = result.Insert(value)
		}
	}
	return result
}

// Difference returns the set of elements of s that are not in other. When s is
// the smaller set its elements are filtered; otherwise the elements of other are
// deleted from s, which shares whatever they do not touch.
// O(m log32 n), m being the smaller size.
func (s Set[T]) Difference(other Set[T]) Set[T] {
	if s.Len() <= other.Len() {
		result := EmptySet[T]()
		for value := range s.All() {
			if !other.Contains(value) {
				result = result.Insert(value)
			}
		}
		return result
	}
	result := s
	for value := range other.All() {
		result = result.Delete(value)
	}
	return result
}

// All returns an iterator over the elements, in an unspecified order. O(n) for a
// full walk.
func (s Set[T]) All() iter.Seq[T] { return s.entries.Keys() }

// Map returns the set of mapper applied to every element. Elements that map to
// the same result collapse into one, so the result can be smaller than the
// receiver, as it can for Tree. O(n log32 n).
func (s Set[T]) Map[U comparable](mapper func(T) U) Set[U] {
	result := EmptySet[U]()
	for value := range s.All() {
		result = result.Insert(mapper(value))
	}
	return result
}

// Filter returns a set of the elements matching the predicate. O(n log32 n).
func (s Set[T]) Filter(predicate func(T) bool) Set[T] {
	result := EmptySet[T]()
	for value := range s.All() {
		if predicate(value) {
			result = result.Insert(value)
		}
	}
	return result
}

// Fold combines the elements, in the order All yields them, starting from
// initial. That order is unspecified, so combine should not depend on it. O(n).
func (s Set[T]) Fold[U any](initial U, combine func(U, T) U) U {
	result := initial
	for value := range s.All() {
		result = combine(result, value)
	}
	return result
}

// String renders the set as Set(a, b, c), elements sorted by their rendering so
// that printing is deterministic even though iterating is not. Elements only
// need to be comparable, so there is no order to sort them by other than their
// text: numbers sort as text too, 10 before 2. O(n log n).
func (s Set[T]) String() string {
	rendered := make([]string, 0, s.Len())
	for value := range s.All() {
		rendered = append(rendered, fmt.Sprint(value))
	}
	slices.Sort(rendered)
	return "Set(" + strings.Join(rendered, ", ") + ")"
}

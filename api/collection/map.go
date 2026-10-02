package collection

import (
	"fmt"
	"iter"
	"maps"
	"strings"

	"github.com/glours/go2funk/api/control"
)

// Map is an immutable, persistent hash map, built on a CHAMP trie.
//
// Keys only need to be comparable, not ordered: structs, arrays and pointers all
// work, hashed with hash/maphash. That is what sets it apart from Tree, which
// keeps its values sorted but needs cmp.Ordered.
//
// Putting or deleting a key rebuilds only the path from the root to it — a
// handful of nodes, since each level of the trie branches 32 ways — and shares
// every other node with the map it came from, so keeping old versions is cheap.
//
// Iteration order is unspecified. A given map yields the same order every time
// it is walked, but two maps holding the same entries may not, and the order
// changes from one process to the next, because the hash seed does.
//
// Keys behave as they do in Go's own map: a float NaN is not equal to itself, so
// an entry keyed by one can be added but never read back or deleted.
//
// The zero value is the empty map. Len and IsEmpty are O(1); Get, ContainsKey,
// Put and Delete are O(log32 n), which no realistic map takes past seven levels;
// anything that visits every entry is O(n).
type Map[K comparable, V any] struct {
	root *champNode[K, V]
}

// EmptyMap returns the empty map. It is the same as the zero value of Map.
func EmptyMap[K comparable, V any]() Map[K, V] {
	return Map[K, V]{}
}

// CollectMap drains seq into a new map. When a key appears more than once, the
// last value wins, as with repeated assignment. O(n log32 n).
//
// It is also the way in from a Go map: CollectMap(maps.All(goMap)). The way back
// out is maps.Collect(m.All()).
func CollectMap[K comparable, V any](seq iter.Seq2[K, V]) Map[K, V] {
	result := EmptyMap[K, V]()
	for key, value := range seq {
		result = result.Put(key, value)
	}
	return result
}

// IsEmpty reports whether the map holds no entry. O(1).
func (m Map[K, V]) IsEmpty() bool { return m.root == nil }

// Len returns the number of entries. O(1).
func (m Map[K, V]) Len() int { return champCount(m.root) }

// Get returns the value bound to key, or None if there is none. O(log32 n).
func (m Map[K, V]) Get(key K) control.Option[V] {
	return control.FromTuple(champGet(m.root, key, hashKey(key), 0))
}

// ContainsKey reports whether key is bound to a value. O(log32 n).
func (m Map[K, V]) ContainsKey(key K) bool {
	_, found := champGet(m.root, key, hashKey(key), 0)
	return found
}

// Put returns a map binding key to value, replacing any value it was bound to.
// O(log32 n).
func (m Map[K, V]) Put(key K, value V) Map[K, V] {
	root, _ := champPut(m.root, key, value, hashKey(key), 0)
	return Map[K, V]{root: root}
}

// Delete returns a map without key. Deleting a key that is not there returns the
// very same map, sharing every node. O(log32 n).
func (m Map[K, V]) Delete(key K) Map[K, V] {
	root, _ := champRemove(m.root, key, hashKey(key), 0)
	return Map[K, V]{root: root}
}

// All returns an iterator over the entries, in an unspecified order. O(n) for a
// full walk.
func (m Map[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) { champAll(m.root, yield) }
}

// Keys returns an iterator over the keys, in the same order as All.
func (m Map[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		champAll(m.root, func(key K, _ V) bool { return yield(key) })
	}
}

// Values returns an iterator over the values, in the same order as All.
func (m Map[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		champAll(m.root, func(_ K, value V) bool { return yield(value) })
	}
}

// Map returns the map binding every key to mapper applied to its value.
//
// This is Vavr's mapValues rather than its map, which maps whole entries and
// can merge keys: mapping only the values is what keeps Map a functor over V.
// No key changes, so the result has the receiver's exact shape and nothing is
// hashed again. O(n).
func (m Map[K, V]) Map[U any](mapper func(V) U) Map[K, U] {
	return Map[K, U]{root: champMapValues(m.root, mapper)}
}

// Filter returns a map of the entries matching the predicate. O(n log32 n).
func (m Map[K, V]) Filter(predicate func(K, V) bool) Map[K, V] {
	result := EmptyMap[K, V]()
	for key, value := range m.All() {
		if predicate(key, value) {
			result = result.Put(key, value)
		}
	}
	return result
}

// Fold combines the entries, in the order All yields them, starting from
// initial. That order is unspecified, so combine should not depend on it. O(n).
func (m Map[K, V]) Fold[U any](initial U, combine func(U, K, V) U) U {
	result := initial
	for key, value := range m.All() {
		result = combine(result, key, value)
	}
	return result
}

// String renders the map as Map[a:1 b:2], keys sorted the way fmt sorts a Go
// map's. Printing is therefore deterministic even though iterating is not, at
// the cost of building a Go map first. O(n log n).
func (m Map[K, V]) String() string {
	return "Map" + strings.TrimPrefix(fmt.Sprint(maps.Collect(m.All())), "map")
}

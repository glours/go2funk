package collection_test

import (
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"testing"

	"github.com/glours/go2funk/api/collection"
)

// lookup unwraps Get so a test can compare the value and its presence at once.
func lookup[K comparable, V any](m collection.Map[K, V], key K) (V, bool) {
	return m.Get(key).Get()
}

func TestMapZeroValueIsEmpty(t *testing.T) {
	var m collection.Map[string, int]

	if !m.IsEmpty() || m.Len() != 0 {
		t.Error("the zero value of Map must be the empty map")
	}
	if !m.Get("a").IsEmpty() || m.ContainsKey("a") {
		t.Error("an empty map holds no key")
	}
	if !m.Delete("a").IsEmpty() {
		t.Error("deleting from an empty map gives an empty map")
	}
	for key, value := range m.All() {
		t.Errorf("an empty map yielded %v -> %v", key, value)
	}

	if got, ok := lookup(m.Put("a", 1), "a"); !ok || got != 1 {
		t.Errorf("Put on the zero value: Get(a) = %d, %t, want 1, true", got, ok)
	}
}

func TestMapPutAndGet(t *testing.T) {
	m := collection.EmptyMap[string, int]().Put("one", 1).Put("two", 2).Put("three", 3)

	tests := []struct {
		name  string
		key   string
		want  int
		found bool
	}{
		{name: "first key", key: "one", want: 1, found: true},
		{name: "middle key", key: "two", want: 2, found: true},
		{name: "last key", key: "three", want: 3, found: true},
		{name: "absent key", key: "four", want: 0, found: false},
		{name: "empty string is just another key", key: "", want: 0, found: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := lookup(m, test.key)
			if got != test.want || ok != test.found {
				t.Errorf("Get(%q) = %d, %t, want %d, %t", test.key, got, ok, test.want, test.found)
			}
			if m.ContainsKey(test.key) != test.found {
				t.Errorf("ContainsKey(%q) = %t, want %t", test.key, !test.found, test.found)
			}
		})
	}
	if got := m.Len(); got != 3 {
		t.Errorf("Len = %d, want 3", got)
	}
}

func TestMapPutReplacesTheValue(t *testing.T) {
	m := collection.EmptyMap[string, int]().Put("a", 1)

	replaced := m.Put("a", 2)
	if got, _ := lookup(replaced, "a"); got != 2 {
		t.Errorf("Get(a) after replacing = %d, want 2", got)
	}
	if got := replaced.Len(); got != 1 {
		t.Errorf("replacing a value changed Len to %d, want 1", got)
	}
	if got, _ := lookup(m, "a"); got != 1 {
		t.Errorf("the original map was modified: Get(a) = %d, want 1", got)
	}
}

func TestMapDelete(t *testing.T) {
	m := collection.EmptyMap[string, int]().Put("a", 1).Put("b", 2).Put("c", 3)

	without := m.Delete("b")
	if want := map[string]int{"a": 1, "c": 3}; !maps.Equal(maps.Collect(without.All()), want) {
		t.Errorf("Delete(b) = %v, want %v", maps.Collect(without.All()), want)
	}
	if want := map[string]int{"a": 1, "b": 2, "c": 3}; !maps.Equal(maps.Collect(m.All()), want) {
		t.Errorf("the original map was modified: %v", maps.Collect(m.All()))
	}
	if got := m.Delete("z").Len(); got != 3 {
		t.Errorf("deleting an absent key changed Len to %d, want 3", got)
	}
}

func TestMapDeleteEveryKey(t *testing.T) {
	m := collection.EmptyMap[int, int]()
	for key := range 500 {
		m = m.Put(key, key*key)
	}

	for key := range 500 {
		m = m.Delete(key)
		if want := 500 - key - 1; m.Len() != want {
			t.Fatalf("after deleting %d, Len = %d, want %d", key, m.Len(), want)
		}
		if m.ContainsKey(key) {
			t.Fatalf("%d is still present after Delete", key)
		}
	}
	if !m.IsEmpty() {
		t.Error("the map must be empty once every key is deleted")
	}
}

// Keys are comparable, not ordered: a struct works, which Tree cannot offer.
func TestMapAcceptsStructKeys(t *testing.T) {
	type point struct{ x, y int }

	m := collection.EmptyMap[point, string]().Put(point{1, 2}, "a").Put(point{2, 1}, "b")

	if got, _ := lookup(m, point{1, 2}); got != "a" {
		t.Errorf("Get({1 2}) = %q, want a", got)
	}
	if got, _ := lookup(m, point{2, 1}); got != "b" {
		t.Errorf("Get({2 1}) = %q, want b", got)
	}
	if m.ContainsKey(point{1, 1}) {
		t.Error("ContainsKey({1 1}) = true, want false")
	}
}

// A NaN is not equal to itself, so like a key of Go's own map it can be written
// but never read back. Both sides are asserted, so the parallel is pinned.
func TestMapTreatsNaNLikeAGoMap(t *testing.T) {
	builtin := map[float64]int{}
	builtin[math.NaN()] = 1
	builtin[math.NaN()] = 2
	_, builtinFound := builtin[math.NaN()]

	m := collection.EmptyMap[float64, int]().Put(math.NaN(), 1).Put(math.NaN(), 2)

	if m.Len() != len(builtin) {
		t.Errorf("Len = %d, the built-in map holds %d", m.Len(), len(builtin))
	}
	if m.ContainsKey(math.NaN()) != builtinFound {
		t.Errorf("ContainsKey(NaN) = %t, the built-in map finds it: %t", m.ContainsKey(math.NaN()), builtinFound)
	}
}

func TestCollectMap(t *testing.T) {
	source := map[string]int{"a": 1, "b": 2, "c": 3}

	m := collection.CollectMap(maps.All(source))
	if got := maps.Collect(m.All()); !maps.Equal(got, source) {
		t.Errorf("CollectMap then All = %v, want %v", got, source)
	}

	// A sequence may repeat a key; the last value wins, as with assignment.
	repeated := func(yield func(string, int) bool) {
		_ = yield("a", 1) && yield("b", 2) && yield("a", 3)
	}
	if got := maps.Collect(collection.CollectMap(repeated).All()); !maps.Equal(got, map[string]int{"a": 3, "b": 2}) {
		t.Errorf("CollectMap with a repeated key = %v, want map[a:3 b:2]", got)
	}

	if !collection.CollectMap(maps.All(map[string]int{})).IsEmpty() {
		t.Error("collecting an empty sequence gives an empty map")
	}
}

func TestMapKeysAndValues(t *testing.T) {
	m := collection.CollectMap(maps.All(map[string]int{"a": 1, "b": 2, "c": 3}))

	if got := slices.Sorted(m.Keys()); !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Errorf("Keys = %v, want [a b c]", got)
	}
	if got := slices.Sorted(m.Values()); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("Values = %v, want [1 2 3]", got)
	}
}

// The three iterators must honour a consumer that stops early.
func TestMapIteratorsStopEarly(t *testing.T) {
	m := collection.EmptyMap[int, int]()
	for key := range 100 {
		m = m.Put(key, key)
	}

	tests := []struct {
		name string
		walk func(stopAfter int) int
	}{
		{name: "All", walk: func(stopAfter int) int {
			seen := 0
			for range m.All() {
				if seen++; seen == stopAfter {
					break
				}
			}
			return seen
		}},
		{name: "Keys", walk: func(stopAfter int) int {
			seen := 0
			for range m.Keys() {
				if seen++; seen == stopAfter {
					break
				}
			}
			return seen
		}},
		{name: "Values", walk: func(stopAfter int) int {
			seen := 0
			for range m.Values() {
				if seen++; seen == stopAfter {
					break
				}
			}
			return seen
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.walk(3); got != 3 {
				t.Errorf("visited %d entries after breaking at 3", got)
			}
		})
	}
}

// Iteration order is unspecified, but a given map must yield the same order
// every time it is walked: nothing in it changes between two walks.
func TestMapIterationIsStableForAGivenMap(t *testing.T) {
	m := collection.EmptyMap[int, int]()
	for key := range 200 {
		m = m.Put(key, key)
	}

	first := slices.Collect(m.Keys())
	if second := slices.Collect(m.Keys()); !slices.Equal(first, second) {
		t.Error("walking the same map twice gave two different orders")
	}
}

// Put rebuilds only the path to the key, and a Delete of an absent key hands
// back the very same map. Both are measured, not assumed.
func TestMapSharesStructure(t *testing.T) {
	m := collection.EmptyMap[int, int]()
	for key := range 10000 {
		m = m.Put(key*2, key)
	}

	var kept collection.Map[int, int]

	// 10000 keys sit three or four levels deep in a 32-way trie, and each level
	// rebuilt costs a node and its slices: a count in the tens, not thousands.
	put := testing.AllocsPerRun(50, func() { kept = m.Put(1, 1) })
	t.Logf("putting into a map of 10000 allocates %.0f times", put)
	if put > 20 {
		t.Errorf("putting into a map of 10000 allocated %.0f times, want a shallow count", put)
	}
	if kept.Len() != 10001 || m.Len() != 10000 {
		t.Errorf("Len drifted: new=%d old=%d", kept.Len(), m.Len())
	}

	deleted := testing.AllocsPerRun(50, func() { kept = m.Delete(1) })
	if deleted != 0 {
		t.Errorf("deleting an absent key allocated %.0f times, want 0", deleted)
	}
	if kept.Len() != m.Len() {
		t.Errorf("Len changed to %d, want %d", kept.Len(), m.Len())
	}
}

func TestMapMapChangesTheValueType(t *testing.T) {
	m := collection.CollectMap(maps.All(map[string]int{"a": 1, "b": 2, "c": 3}))

	var labelled collection.Map[string, string] = m.Map(strconv.Itoa)
	if want := map[string]string{"a": "1", "b": "2", "c": "3"}; !maps.Equal(maps.Collect(labelled.All()), want) {
		t.Errorf("Map = %v, want %v", maps.Collect(labelled.All()), want)
	}

	// Keys are untouched, so nothing can collapse: the size never changes.
	constant := m.Map(func(int) bool { return true })
	if got := constant.Len(); got != 3 {
		t.Errorf("mapping every value to the same result changed Len to %d, want 3", got)
	}

	var zero collection.Map[string, int]
	if !zero.Map(strconv.Itoa).IsEmpty() {
		t.Error("mapping an empty map gives an empty map")
	}
}

// The mapped map is a full map in its own right, not a view on the original.
func TestMapMapResultIsIndependent(t *testing.T) {
	m := collection.EmptyMap[int, int]()
	for key := range 300 {
		m = m.Put(key, key)
	}

	doubled := m.Map(func(value int) int { return value * 2 }).Put(1000, 0).Delete(0)

	if got, _ := lookup(doubled, 150); got != 300 {
		t.Errorf("Get(150) on the mapped map = %d, want 300", got)
	}
	if doubled.Len() != 300 || doubled.ContainsKey(0) || !doubled.ContainsKey(1000) {
		t.Errorf("the mapped map did not take Put and Delete: Len = %d", doubled.Len())
	}
	if got, _ := lookup(m, 150); got != 150 || m.Len() != 300 {
		t.Errorf("the original map was modified: Get(150) = %d, Len = %d", got, m.Len())
	}
}

func TestMapFilter(t *testing.T) {
	m := collection.CollectMap(maps.All(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}))

	tests := []struct {
		name      string
		predicate func(string, int) bool
		want      map[string]int
	}{
		{name: "by value", predicate: func(_ string, value int) bool { return value%2 == 0 }, want: map[string]int{"b": 2, "d": 4}},
		{name: "by key", predicate: func(key string, _ int) bool { return key < "c" }, want: map[string]int{"a": 1, "b": 2}},
		{name: "keeps everything", predicate: func(string, int) bool { return true }, want: map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}},
		{name: "keeps nothing", predicate: func(string, int) bool { return false }, want: map[string]int{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filtered := m.Filter(test.predicate)
			if got := maps.Collect(filtered.All()); !maps.Equal(got, test.want) {
				t.Errorf("Filter = %v, want %v", got, test.want)
			}
			if filtered.Len() != len(test.want) {
				t.Errorf("Len = %d, want %d", filtered.Len(), len(test.want))
			}
		})
	}
	if m.Len() != 4 {
		t.Errorf("the original map was modified: Len = %d", m.Len())
	}
}

func TestMapFold(t *testing.T) {
	m := collection.CollectMap(maps.All(map[string]int{"a": 1, "bb": 2, "ccc": 3}))

	sum := m.Fold(0, func(acc int, _ string, value int) int { return acc + value })
	if sum != 6 {
		t.Errorf("sum of values = %d, want 6", sum)
	}

	// The accumulator can be of a third type, unrelated to K and V.
	lengths := m.Fold([]int{}, func(acc []int, key string, _ int) []int { return append(acc, len(key)) })
	if slices.Sort(lengths); !slices.Equal(lengths, []int{1, 2, 3}) {
		t.Errorf("key lengths = %v, want [1 2 3]", lengths)
	}

	var zero collection.Map[string, int]
	if got := zero.Fold(42, func(int, string, int) int { return 0 }); got != 42 {
		t.Errorf("folding an empty map = %d, want the initial 42", got)
	}
}

// String sorts by key, as fmt does for a Go map, so printing a Map is
// deterministic even though iterating over it is not.
func TestMapString(t *testing.T) {
	tests := []struct {
		name string
		m    fmt.Stringer
		want string
	}{
		{name: "empty", m: collection.EmptyMap[string, int](), want: "Map[]"},
		{name: "string keys", m: collection.CollectMap(maps.All(map[string]int{"c": 3, "a": 1, "b": 2})), want: "Map[a:1 b:2 c:3]"},
		{name: "int keys", m: collection.CollectMap(maps.All(map[int]string{10: "x", 2: "y", 33: "z"})), want: "Map[2:y 10:x 33:z]"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.m.String(); got != test.want {
				t.Errorf("String = %q, want %q", got, test.want)
			}
		})
	}
}

// Put on a persistent map costs the depth of the trie, not its size: while the
// map grows a thousandfold, allocations per Put go from about six to about ten,
// one more level of nodes each time the map grows thirty-twofold. Time grows
// faster than that, because a large trie no longer fits in the CPU cache.
func BenchmarkMapPut(b *testing.B) {
	for _, size := range []int{1_000, 100_000, 1_000_000} {
		m := collection.EmptyMap[int, int]()
		for key := range size {
			m = m.Put(key, key)
		}
		b.Run(fmt.Sprintf("persistent/%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; b.Loop(); i++ {
				_ = m.Put(size+i, i)
			}
		})
	}
}

func BenchmarkMapGet(b *testing.B) {
	for _, size := range []int{1_000, 100_000, 1_000_000} {
		m := collection.EmptyMap[int, int]()
		for key := range size {
			m = m.Put(key, key)
		}
		b.Run(fmt.Sprintf("persistent/%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; b.Loop(); i++ {
				_ = m.Get(i % size)
			}
		})
	}
}

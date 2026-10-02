package collection_test

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"testing"

	"github.com/glours/go2funk/api/collection"
)

// sorted reads a set's contents in a fixed order, since iteration has none.
func sorted[T cmp.Ordered](s collection.Set[T]) []T {
	return slices.Sorted(s.All())
}

// rangeSet builds the set {from, ..., to-1}.
func rangeSet(from, to int) collection.Set[int] {
	s := collection.EmptySet[int]()
	for value := from; value < to; value++ {
		s = s.Insert(value)
	}
	return s
}

func TestSetZeroValueIsEmpty(t *testing.T) {
	var s collection.Set[string]

	if !s.IsEmpty() || s.Len() != 0 {
		t.Error("the zero value of Set must be the empty set")
	}
	if s.Contains("a") {
		t.Error("an empty set contains nothing")
	}
	if !s.Delete("a").IsEmpty() {
		t.Error("deleting from an empty set gives an empty set")
	}
	for value := range s.All() {
		t.Errorf("an empty set yielded %q", value)
	}
	if got := sorted(s.Insert("a")); !slices.Equal(got, []string{"a"}) {
		t.Errorf("Insert on the zero value = %v, want [a]", got)
	}
	if got := sorted(s.Union(collection.SetOf("b"))); !slices.Equal(got, []string{"b"}) {
		t.Errorf("Union with the zero value = %v, want [b]", got)
	}
}

func TestSetOfCollapsesDuplicates(t *testing.T) {
	s := collection.SetOf(3, 1, 2, 3, 1)

	if got := sorted(s); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("SetOf = %v, want [1 2 3]", got)
	}
	if got := s.Len(); got != 3 {
		t.Errorf("Len = %d, want 3", got)
	}
}

func TestSetContains(t *testing.T) {
	s := collection.SetOf("a", "b", "c")

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "first", value: "a", want: true},
		{name: "last", value: "c", want: true},
		{name: "absent", value: "d", want: false},
		{name: "empty string is just another value", value: "", want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := s.Contains(test.value); got != test.want {
				t.Errorf("Contains(%q) = %t, want %t", test.value, got, test.want)
			}
		})
	}
}

func TestSetInsertAndDelete(t *testing.T) {
	s := collection.SetOf(1, 2, 3)

	if got := sorted(s.Insert(4)); !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Errorf("Insert(4) = %v, want [1 2 3 4]", got)
	}
	if got := sorted(s.Insert(2)); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("inserting an existing value = %v, want [1 2 3]", got)
	}
	if got := sorted(s.Delete(2)); !slices.Equal(got, []int{1, 3}) {
		t.Errorf("Delete(2) = %v, want [1 3]", got)
	}
	if got := sorted(s.Delete(9)); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("deleting an absent value = %v, want [1 2 3]", got)
	}
	if got := sorted(s); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("the original set was modified: %v", got)
	}
}

// Elements are comparable, not ordered: a struct works, which Tree cannot offer.
func TestSetAcceptsStructElements(t *testing.T) {
	type point struct{ x, y int }

	s := collection.SetOf(point{1, 2}, point{2, 1}, point{1, 2})

	if s.Len() != 2 || !s.Contains(point{1, 2}) || !s.Contains(point{2, 1}) || s.Contains(point{1, 1}) {
		t.Errorf("a set of structs holds the wrong elements: Len = %d", s.Len())
	}
}

func TestCollectSet(t *testing.T) {
	s := collection.CollectSet(slices.Values([]string{"b", "a", "b"}))

	if got := sorted(s); !slices.Equal(got, []string{"a", "b"}) {
		t.Errorf("CollectSet = %v, want [a b]", got)
	}
	if !collection.CollectSet(slices.Values([]string{})).IsEmpty() {
		t.Error("collecting an empty sequence gives an empty set")
	}
}

func TestSetUnion(t *testing.T) {
	tests := []struct {
		name        string
		left, right collection.Set[int]
		want        []int
	}{
		{name: "overlapping", left: collection.SetOf(1, 2, 3), right: collection.SetOf(3, 4), want: []int{1, 2, 3, 4}},
		{name: "smaller on the left", left: collection.SetOf(3, 4), right: collection.SetOf(1, 2, 3), want: []int{1, 2, 3, 4}},
		{name: "disjoint", left: collection.SetOf(1), right: collection.SetOf(2), want: []int{1, 2}},
		{name: "identical", left: collection.SetOf(1, 2), right: collection.SetOf(1, 2), want: []int{1, 2}},
		{name: "empty on the left", left: collection.EmptySet[int](), right: collection.SetOf(1, 2), want: []int{1, 2}},
		{name: "empty on the right", left: collection.SetOf(1, 2), right: collection.EmptySet[int](), want: []int{1, 2}},
		{name: "both empty", left: collection.EmptySet[int](), right: collection.EmptySet[int](), want: []int{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			union := test.left.Union(test.right)
			if got := sorted(union); !slices.Equal(got, test.want) {
				t.Errorf("Union = %v, want %v", got, test.want)
			}
			if union.Len() != len(test.want) {
				t.Errorf("Len = %d, want %d", union.Len(), len(test.want))
			}
		})
	}
}

func TestSetIntersection(t *testing.T) {
	tests := []struct {
		name        string
		left, right collection.Set[int]
		want        []int
	}{
		{name: "overlapping", left: collection.SetOf(1, 2, 3, 4), right: collection.SetOf(3, 4, 5), want: []int{3, 4}},
		{name: "smaller on the left", left: collection.SetOf(3, 4, 5), right: collection.SetOf(1, 2, 3, 4), want: []int{3, 4}},
		{name: "disjoint", left: collection.SetOf(1), right: collection.SetOf(2), want: []int{}},
		{name: "subset", left: collection.SetOf(1, 2, 3), right: collection.SetOf(2), want: []int{2}},
		{name: "empty on the right", left: collection.SetOf(1, 2), right: collection.EmptySet[int](), want: []int{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			intersection := test.left.Intersection(test.right)
			if got := sorted(intersection); !slices.Equal(got, test.want) {
				t.Errorf("Intersection = %v, want %v", got, test.want)
			}
			if intersection.Len() != len(test.want) {
				t.Errorf("Len = %d, want %d", intersection.Len(), len(test.want))
			}
		})
	}
}

// Difference is the one operation that is not symmetric, so both sides of the
// size comparison are covered with the same contents swapped.
func TestSetDifference(t *testing.T) {
	tests := []struct {
		name        string
		left, right collection.Set[int]
		want        []int
	}{
		{name: "larger minus smaller", left: collection.SetOf(1, 2, 3, 4), right: collection.SetOf(2, 9), want: []int{1, 3, 4}},
		{name: "smaller minus larger", left: collection.SetOf(2, 9), right: collection.SetOf(1, 2, 3, 4), want: []int{9}},
		{name: "disjoint", left: collection.SetOf(1, 2), right: collection.SetOf(3), want: []int{1, 2}},
		{name: "identical", left: collection.SetOf(1, 2), right: collection.SetOf(1, 2), want: []int{}},
		{name: "minus empty", left: collection.SetOf(1, 2), right: collection.EmptySet[int](), want: []int{1, 2}},
		{name: "empty minus", left: collection.EmptySet[int](), right: collection.SetOf(1, 2), want: []int{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			difference := test.left.Difference(test.right)
			if got := sorted(difference); !slices.Equal(got, test.want) {
				t.Errorf("Difference = %v, want %v", got, test.want)
			}
			if difference.Len() != len(test.want) {
				t.Errorf("Len = %d, want %d", difference.Len(), len(test.want))
			}
		})
	}
}

func TestSetOperationsLeaveTheOperandsUntouched(t *testing.T) {
	left, right := collection.SetOf(1, 2, 3), collection.SetOf(3, 4)

	left.Union(right)
	left.Intersection(right)
	left.Difference(right)
	right.Difference(left)

	if got := sorted(left); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("the left operand was modified: %v", got)
	}
	if got := sorted(right); !slices.Equal(got, []int{3, 4}) {
		t.Errorf("the right operand was modified: %v", got)
	}
}

func TestSetAllStopsEarly(t *testing.T) {
	seen := 0
	for range rangeSet(0, 100).All() {
		if seen++; seen == 3 {
			break
		}
	}
	if seen != 3 {
		t.Errorf("visited %d values after breaking at 3", seen)
	}
}

// A no-op Insert or Delete hands back the very same set, as it does for Tree.
func TestSetNoOpSharesEverything(t *testing.T) {
	s := rangeSet(0, 10000)

	var kept collection.Set[int]

	inserted := testing.AllocsPerRun(50, func() { kept = s.Insert(42) })
	if inserted != 0 {
		t.Errorf("re-inserting an existing value allocated %.0f times, want 0", inserted)
	}
	if kept.Len() != s.Len() {
		t.Errorf("Len changed to %d, want %d", kept.Len(), s.Len())
	}

	deleted := testing.AllocsPerRun(50, func() { kept = s.Delete(-1) })
	if deleted != 0 {
		t.Errorf("deleting an absent value allocated %.0f times, want 0", deleted)
	}
	if kept.Len() != s.Len() {
		t.Errorf("Len changed to %d, want %d", kept.Len(), s.Len())
	}
}

// Union and Difference walk the smaller operand, so combining a handful of
// values with a large set costs a handful of trie operations whichever side it
// is on. Walking the large side instead would cost thousands of allocations.
//
// Intersection walks the smaller side too, but allocations cannot see it: the
// walk only looks values up, which allocates nothing either way. That one is
// shown by BenchmarkSetIntersection instead.
func TestSetOperationsWalkTheSmallerSide(t *testing.T) {
	large, small := rangeSet(0, 10000), collection.SetOf(5, 20000, 30000)

	tests := []struct {
		name      string
		operation func() collection.Set[int]
	}{
		{name: "Union, small on the right", operation: func() collection.Set[int] { return large.Union(small) }},
		{name: "Union, small on the left", operation: func() collection.Set[int] { return small.Union(large) }},
		{name: "Difference, small on the right", operation: func() collection.Set[int] { return large.Difference(small) }},
		{name: "Difference, small on the left", operation: func() collection.Set[int] { return small.Difference(large) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Three values, each costing at most one path rebuilt: well under 50.
			if allocations := testing.AllocsPerRun(20, func() { test.operation() }); allocations > 50 {
				t.Errorf("allocated %.0f times, want a count bounded by the small side", allocations)
			}
		})
	}
}

// Intersection costs the size of the smaller operand, whichever side it is on:
// the two orders below should run in about the same time, and neither should
// grow with the large set.
func BenchmarkSetIntersection(b *testing.B) {
	small := collection.SetOf(5, 20000, 30000)
	for _, size := range []int{1_000, 100_000} {
		large := rangeSet(0, size)
		b.Run(fmt.Sprintf("large-first/%d", size), func(b *testing.B) {
			for b.Loop() {
				_ = large.Intersection(small)
			}
		})
		b.Run(fmt.Sprintf("small-first/%d", size), func(b *testing.B) {
			for b.Loop() {
				_ = small.Intersection(large)
			}
		})
	}
}

func TestSetMap(t *testing.T) {
	s := collection.SetOf(1, 2, 3)

	var labelled collection.Set[string] = s.Map(strconv.Itoa)
	if got := sorted(labelled); !slices.Equal(got, []string{"1", "2", "3"}) {
		t.Errorf("Map = %v, want [1 2 3]", got)
	}

	// Values that map to the same result collapse: a set holds each once.
	parity := collection.SetOf(1, 2, 3, 4).Map(func(value int) int { return value % 2 })
	if got := sorted(parity); !slices.Equal(got, []int{0, 1}) {
		t.Errorf("Map = %v, want [0 1]", got)
	}

	var zero collection.Set[int]
	if !zero.Map(strconv.Itoa).IsEmpty() {
		t.Error("mapping an empty set gives an empty set")
	}
}

func TestSetFilter(t *testing.T) {
	s := rangeSet(0, 10)

	tests := []struct {
		name      string
		predicate func(int) bool
		want      []int
	}{
		{name: "even", predicate: func(value int) bool { return value%2 == 0 }, want: []int{0, 2, 4, 6, 8}},
		{name: "keeps everything", predicate: func(int) bool { return true }, want: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{name: "keeps nothing", predicate: func(int) bool { return false }, want: []int{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := sorted(s.Filter(test.predicate)); !slices.Equal(got, test.want) {
				t.Errorf("Filter = %v, want %v", got, test.want)
			}
		})
	}
	if s.Len() != 10 {
		t.Errorf("the original set was modified: Len = %d", s.Len())
	}
}

func TestSetFold(t *testing.T) {
	s := collection.SetOf(1, 2, 3, 4)

	if got := s.Fold(0, func(sum, value int) int { return sum + value }); got != 10 {
		t.Errorf("sum = %d, want 10", got)
	}

	// The accumulator can be of another type than the elements.
	labels := s.Fold([]string{}, func(acc []string, value int) []string { return append(acc, strconv.Itoa(value)) })
	if slices.Sort(labels); !slices.Equal(labels, []string{"1", "2", "3", "4"}) {
		t.Errorf("labels = %v, want [1 2 3 4]", labels)
	}

	var zero collection.Set[int]
	if got := zero.Fold(42, func(int, int) int { return 0 }); got != 42 {
		t.Errorf("folding an empty set = %d, want the initial 42", got)
	}
}

// String sorts the elements by their rendering, so printing a Set is
// deterministic even though iterating over it is not. Numbers therefore sort as
// text, which the last case pins rather than hides.
func TestSetString(t *testing.T) {
	tests := []struct {
		name string
		s    fmt.Stringer
		want string
	}{
		{name: "empty", s: collection.EmptySet[string](), want: "Set()"},
		{name: "strings", s: collection.SetOf("c", "a", "b"), want: "Set(a, b, c)"},
		{name: "structs", s: collection.SetOf(struct{ x, y int }{2, 1}, struct{ x, y int }{1, 2}), want: "Set({1 2}, {2 1})"},
		{name: "numbers sort as text", s: collection.SetOf(2, 10, 1), want: "Set(1, 10, 2)"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.s.String(); got != test.want {
				t.Errorf("String = %q, want %q", got, test.want)
			}
		})
	}
}

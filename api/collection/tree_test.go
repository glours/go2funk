package collection_test

import (
	"iter"
	"math"
	"slices"
	"strconv"
	"testing"

	"github.com/glours/go2funk/api/collection"
)

func TestTreeZeroValueIsEmpty(t *testing.T) {
	var tree collection.Tree[int]

	if !tree.IsEmpty() || tree.Len() != 0 {
		t.Error("the zero value of Tree must be the empty tree")
	}
	if !tree.Min().IsEmpty() || !tree.Max().IsEmpty() {
		t.Error("an empty tree has no minimum and no maximum")
	}
	if tree.Contains(1) {
		t.Error("an empty tree contains nothing")
	}
	if got := tree.ToSlice(); len(got) != 0 {
		t.Errorf("ToSlice = %v, want []", got)
	}
}

func TestTreeKeepsValuesSorted(t *testing.T) {
	tree := collection.TreeOf(5, 3, 8, 1, 9, 7, 2, 6, 4)

	want := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	if got := tree.ToSlice(); !slices.Equal(got, want) {
		t.Errorf("ToSlice = %v, want %v", got, want)
	}
	if got := tree.Len(); got != 9 {
		t.Errorf("Len = %d, want 9", got)
	}
}

func TestTreeInsertIsIdempotent(t *testing.T) {
	tree := collection.TreeOf(1, 2, 3)

	again := tree.Insert(2)
	if got := again.Len(); got != 3 {
		t.Errorf("inserting an existing value changed Len to %d, want 3", got)
	}
	if got := again.ToSlice(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("ToSlice = %v, want [1 2 3]", got)
	}
}

func TestTreeDelete(t *testing.T) {
	tree := collection.TreeOf(5, 3, 8, 1, 9)

	without := tree.Delete(3)
	if want := []int{1, 5, 8, 9}; !slices.Equal(without.ToSlice(), want) {
		t.Errorf("Delete(3) = %v, want %v", without.ToSlice(), want)
	}
	if got := tree.ToSlice(); !slices.Equal(got, []int{1, 3, 5, 8, 9}) {
		t.Errorf("the original tree was modified: %v", got)
	}

	if got := tree.Delete(42).Len(); got != 5 {
		t.Errorf("deleting an absent value changed Len to %d, want 5", got)
	}

	var empty collection.Tree[int]
	if !empty.Delete(1).IsEmpty() {
		t.Error("deleting from an empty tree gives an empty tree")
	}
}

func TestTreeDeleteEveryValue(t *testing.T) {
	values := []int{5, 3, 8, 1, 9, 7, 2, 6, 4}
	tree := collection.TreeOf(values...)

	for i, value := range values {
		tree = tree.Delete(value)
		if want := len(values) - i - 1; tree.Len() != want {
			t.Fatalf("after deleting %d, Len = %d, want %d", value, tree.Len(), want)
		}
		if tree.Contains(value) {
			t.Fatalf("%d is still present after Delete", value)
		}
		if got := tree.ToSlice(); !slices.IsSorted(got) {
			t.Fatalf("the tree lost its order after deleting %d: %v", value, got)
		}
	}
	if !tree.IsEmpty() {
		t.Error("the tree must be empty once every value is deleted")
	}
}

func TestTreeContainsMinMax(t *testing.T) {
	tree := collection.TreeOf(5, 3, 8, 1, 9)

	for _, value := range []int{1, 3, 5, 8, 9} {
		if !tree.Contains(value) {
			t.Errorf("Contains(%d) = false, want true", value)
		}
	}
	for _, value := range []int{0, 4, 10} {
		if tree.Contains(value) {
			t.Errorf("Contains(%d) = true, want false", value)
		}
	}

	if got := tree.Min().OrElse(-1); got != 1 {
		t.Errorf("Min = %d, want 1", got)
	}
	if got := tree.Max().OrElse(-1); got != 9 {
		t.Errorf("Max = %d, want 9", got)
	}
}

func TestTreeIteration(t *testing.T) {
	tree := collection.TreeOf(5, 3, 8, 1, 9)

	var forward []int
	for value := range tree.All() {
		forward = append(forward, value)
	}
	if want := []int{1, 3, 5, 8, 9}; !slices.Equal(forward, want) {
		t.Errorf("All = %v, want %v", forward, want)
	}

	var backward []int
	for value := range tree.Backward() {
		backward = append(backward, value)
	}
	if want := []int{9, 8, 5, 3, 1}; !slices.Equal(backward, want) {
		t.Errorf("Backward = %v, want %v", backward, want)
	}

	// A break must stop the walk, whichever direction it goes.
	for name, seq := range map[string]iter.Seq[int]{
		"All":      tree.All(),
		"Backward": tree.Backward(),
		"Range":    tree.Range(1, 9),
	} {
		count := 0
		for range seq {
			count++
			break
		}
		if count != 1 {
			t.Errorf("%s: break stopped after %d values, want 1", name, count)
		}
	}
}

// Range must stop as soon as the consumer does, at any point of the walk.
func TestTreeRangeStopsEarly(t *testing.T) {
	tree := collection.TreeOf(1, 2, 3, 4, 5, 6, 7, 8, 9)

	for _, stopAfter := range []int{1, 2, 3, 5} {
		var got []int
		for value := range tree.Range(2, 8) {
			got = append(got, value)
			if len(got) == stopAfter {
				break
			}
		}
		if len(got) != stopAfter {
			t.Errorf("stopping after %d gave %v", stopAfter, got)
		}
		if !slices.IsSorted(got) {
			t.Errorf("values came out unordered: %v", got)
		}
	}
}

func TestTreeRange(t *testing.T) {
	tree := collection.TreeOf(1, 3, 5, 7, 9)

	for _, tc := range []struct {
		name     string
		from, to int
		want     []int
	}{
		{"middle", 3, 7, []int{3, 5, 7}},
		{"outside on both ends", 0, 10, []int{1, 3, 5, 7, 9}},
		{"between two values", 4, 4, nil},
		{"single value", 5, 5, []int{5}},
		{"empty range", 7, 3, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got []int
			for value := range tree.Range(tc.from, tc.to) {
				got = append(got, value)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("Range(%d, %d) = %v, want %v", tc.from, tc.to, got, tc.want)
			}
		})
	}
}

func TestTreeFilterAndFold(t *testing.T) {
	tree := collection.TreeOf(1, 2, 3, 4, 5)

	isEven := func(value int) bool { return value%2 == 0 }
	if got := tree.Filter(isEven).ToSlice(); !slices.Equal(got, []int{2, 4}) {
		t.Errorf("Filter = %v, want [2 4]", got)
	}

	sum := tree.Fold(0, func(acc, value int) int { return acc + value })
	if sum != 15 {
		t.Errorf("Fold = %d, want 15", sum)
	}

	var joined string = tree.Fold("", func(acc string, value int) string { return acc + strconv.Itoa(value) })
	if joined != "12345" {
		t.Errorf("Fold to string = %q, want %q", joined, "12345")
	}
}

func TestTreeCollectAndString(t *testing.T) {
	tree := collection.TreeOf(3, 1, 2)

	if got := collection.CollectTree(tree.All()).ToSlice(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("CollectTree = %v, want [1 2 3]", got)
	}
	if got := tree.String(); got != "Tree(1, 2, 3)" {
		t.Errorf("String = %q, want %q", got, "Tree(1, 2, 3)")
	}
	if got := collection.EmptyTree[int]().String(); got != "Tree()" {
		t.Errorf("empty String = %q, want %q", got, "Tree()")
	}
}

func TestTreeWorksWithStrings(t *testing.T) {
	tree := collection.TreeOf("pear", "apple", "fig")

	if got := tree.ToSlice(); !slices.Equal(got, []string{"apple", "fig", "pear"}) {
		t.Errorf("ToSlice = %v, want [apple fig pear]", got)
	}
}

// cmp.Ordered admits floats, and a NaN compares false against everything with
// the < and > operators. Ordering through cmp.Compare instead gives it a place:
// below every other value, and equal to itself.
func TestTreeHandlesNaN(t *testing.T) {
	nan := math.NaN()

	tree := collection.TreeOf(nan, 1.0, 2.0, 3.0)
	if got := tree.Len(); got != 4 {
		t.Errorf("Len = %d, want 4: no value may be swallowed by a NaN", got)
	}
	if got := tree.ToSlice(); len(got) != 4 || !math.IsNaN(got[0]) {
		t.Errorf("ToSlice = %v, want a NaN first then 1 2 3", got)
	}

	// Inserted after the others, it must still find its place.
	other := collection.TreeOf(1.0, 2.0, nan, 3.0)
	if got := other.Len(); got != 4 {
		t.Errorf("Len = %d, want 4: the NaN must not be dropped", got)
	}
	if !other.Contains(nan) {
		t.Error("Contains(NaN) = false: a NaN equals itself under cmp.Compare")
	}

	// A set holds it once.
	if got := collection.TreeOf(nan, nan, nan).Len(); got != 1 {
		t.Errorf("Len = %d, want 1: NaN equals itself", got)
	}
}

// A no-op Insert or Delete must hand back the very same tree, not rebuild the
// path to where the value would have gone.
func TestTreeNoOpSharesEverything(t *testing.T) {
	tree := collection.EmptyTree[int]()
	for value := range 1000 {
		tree = tree.Insert(value * 2)
	}

	var kept collection.Tree[int]

	inserted := testing.AllocsPerRun(50, func() { kept = tree.Insert(1998) })
	if inserted != 0 {
		t.Errorf("re-inserting an existing value allocated %.0f nodes, want 0", inserted)
	}
	if kept.Len() != tree.Len() {
		t.Errorf("Len changed to %d, want %d", kept.Len(), tree.Len())
	}

	deleted := testing.AllocsPerRun(50, func() { kept = tree.Delete(1) })
	if deleted != 0 {
		t.Errorf("deleting an absent value allocated %.0f nodes, want 0", deleted)
	}
	if kept.Len() != tree.Len() {
		t.Errorf("Len changed to %d, want %d", kept.Len(), tree.Len())
	}
}

func TestTreeMap(t *testing.T) {
	tree := collection.TreeOf(1, 2, 3)

	var doubled collection.Tree[string] = tree.Map(strconv.Itoa)
	if got := doubled.ToSlice(); !slices.Equal(got, []string{"1", "2", "3"}) {
		t.Errorf("Map = %v, want [1 2 3]", got)
	}

	// Values that map to the same result collapse: a set holds each once.
	collapsed := collection.TreeOf(1, 2, 3, 4).Map(func(value int) int { return value % 2 })
	if got := collapsed.ToSlice(); !slices.Equal(got, []int{0, 1}) {
		t.Errorf("Map = %v, want [0 1]", got)
	}

	if !collection.EmptyTree[int]().Map(strconv.Itoa).IsEmpty() {
		t.Error("mapping an empty tree gives an empty tree")
	}
}

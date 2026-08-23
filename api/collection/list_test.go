package collection_test

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/glours/go2funk/api/collection"
)

func TestZeroValueIsEmpty(t *testing.T) {
	var l collection.List[int]

	if !l.IsEmpty() || l.Length() != 0 {
		t.Error("the zero value of List must be the empty list")
	}
	if !l.Head().IsEmpty() {
		t.Error("the head of an empty list must be None")
	}
	if got := l.ToSlice(); len(got) != 0 {
		t.Errorf("ToSlice = %v, want []", got)
	}
}

func TestConstructors(t *testing.T) {
	if got := collection.Of(1, 2, 3).ToSlice(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("Of = %v, want [1 2 3]", got)
	}
	if !collection.Of[int]().IsEmpty() {
		t.Error("Of with no argument must be empty")
	}
	if !collection.Empty[int]().IsEmpty() {
		t.Error("Empty must be empty")
	}

	values := []int{1, 2, 3}
	if got := collection.Of(values...).ToSlice(); !slices.Equal(got, values) {
		t.Errorf("Of(slice...) = %v, want %v", got, values)
	}
}

// The list must be readable: this is what the interface-based version could not do.
func TestHeadTailGet(t *testing.T) {
	l := collection.Of(1, 2, 3)

	if got := l.Head().OrElse(-1); got != 1 {
		t.Errorf("Head = %d, want 1", got)
	}
	if got := l.Tail().ToSlice(); !slices.Equal(got, []int{2, 3}) {
		t.Errorf("Tail = %v, want [2 3]", got)
	}
	if got := l.Get(1).OrElse(-1); got != 2 {
		t.Errorf("Get(1) = %d, want 2", got)
	}
	if !l.Get(3).IsEmpty() {
		t.Error("Get out of range must be None")
	}
	if !l.Get(-1).IsEmpty() {
		t.Error("Get with a negative index must be None")
	}
	if !collection.Empty[int]().Tail().IsEmpty() {
		t.Error("the tail of an empty list is the empty list")
	}
}

func TestAllIsAnIterSeq(t *testing.T) {
	l := collection.Of(1, 2, 3)

	var seen []int
	for value := range l.All() {
		seen = append(seen, value)
	}
	if !slices.Equal(seen, []int{1, 2, 3}) {
		t.Errorf("range over All = %v, want [1 2 3]", seen)
	}

	// A break must stop the iteration.
	count := 0
	for range l.All() {
		count++
		break
	}
	if count != 1 {
		t.Errorf("break stopped after %d elements, want 1", count)
	}

	if got := collection.Collect(l.All()).ToSlice(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("Collect = %v, want [1 2 3]", got)
	}
}

func TestPrependAppendAndPersistence(t *testing.T) {
	original := collection.Of(2, 3)

	prepended := original.Prepend(1)
	appended := original.Append(4)

	if got := prepended.ToSlice(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("Prepend = %v, want [1 2 3]", got)
	}
	if got := appended.ToSlice(); !slices.Equal(got, []int{2, 3, 4}) {
		t.Errorf("Append = %v, want [2 3 4]", got)
	}
	if got := original.ToSlice(); !slices.Equal(got, []int{2, 3}) {
		t.Errorf("the original list was mutated: %v, want [2 3]", got)
	}
	if got := original.AppendAll(4, 5).ToSlice(); !slices.Equal(got, []int{2, 3, 4, 5}) {
		t.Errorf("AppendAll = %v, want [2 3 4 5]", got)
	}
}

func TestLength(t *testing.T) {
	for _, tc := range []struct {
		list collection.List[int]
		want int
	}{
		{collection.Empty[int](), 0},
		{collection.Of(1), 1},
		{collection.Of(1, 2, 3), 3},
		{collection.Of(1, 2, 3).Tail(), 2},
		{collection.Of(1, 2, 3).Prepend(0), 4},
	} {
		if got := tc.list.Length(); got != tc.want {
			t.Errorf("Length of %v = %d, want %d", tc.list.ToSlice(), got, tc.want)
		}
	}
}

// The point of the rewrite: Map is a method and changes the element type.
func TestMapChangesTheElementType(t *testing.T) {
	var mapped collection.List[string] = collection.Of(1, 2, 3).Map(strconv.Itoa)

	if got := mapped.ToSlice(); !slices.Equal(got, []string{"1", "2", "3"}) {
		t.Errorf("Map = %v, want [1 2 3]", got)
	}
	if !collection.Empty[int]().Map(strconv.Itoa).IsEmpty() {
		t.Error("mapping an empty list gives an empty list")
	}
}

func TestChaining(t *testing.T) {
	isEven := func(value int) bool { return value%2 == 0 }

	got := collection.Of(1, 2, 3, 4, 5, 6).
		Filter(isEven).
		Map(strconv.Itoa).
		Map(strings.ToUpper).
		ToSlice()

	if !slices.Equal(got, []string{"2", "4", "6"}) {
		t.Errorf("chain = %v, want [2 4 6]", got)
	}
}

func TestFlatMap(t *testing.T) {
	twice := func(value int) collection.List[string] {
		s := strconv.Itoa(value)
		return collection.Of(s, s)
	}

	got := collection.Of(1, 2).FlatMap(twice).ToSlice()
	if !slices.Equal(got, []string{"1", "1", "2", "2"}) {
		t.Errorf("FlatMap = %v, want [1 1 2 2]", got)
	}
}

func TestFilterFoldForEach(t *testing.T) {
	l := collection.Of(1, 2, 3, 4)

	if got := l.Filter(func(v int) bool { return v > 2 }).ToSlice(); !slices.Equal(got, []int{3, 4}) {
		t.Errorf("Filter = %v, want [3 4]", got)
	}

	sum := l.Fold(0, func(acc, value int) int { return acc + value })
	if sum != 10 {
		t.Errorf("Fold = %d, want 10", sum)
	}

	// Fold can change the type too.
	var joined string = l.Fold("", func(acc string, value int) string { return acc + strconv.Itoa(value) })
	if joined != "1234" {
		t.Errorf("Fold to string = %q, want %q", joined, "1234")
	}

	seen := 0
	l.ForEach(func(value int) { seen += value })
	if seen != 10 {
		t.Errorf("ForEach summed %d, want 10", seen)
	}
}

func TestReverse(t *testing.T) {
	if got := collection.Of(1, 2, 3).Reverse().ToSlice(); !slices.Equal(got, []int{3, 2, 1}) {
		t.Errorf("Reverse = %v, want [3 2 1]", got)
	}
	if !collection.Empty[int]().Reverse().IsEmpty() {
		t.Error("reversing an empty list gives an empty list")
	}
}

func TestInsert(t *testing.T) {
	l := collection.Of(1, 2, 4)

	inserted, err := l.Insert(2, 3)
	if err != nil {
		t.Fatalf("Insert returned %v", err)
	}
	if got := inserted.ToSlice(); !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Errorf("Insert = %v, want [1 2 3 4]", got)
	}

	if atEnd, err := l.Insert(3, 9); err != nil || atEnd.Length() != 4 {
		t.Errorf("Insert at the end failed: %v", err)
	}

	// The error must name the caller's index, not the recursion's.
	if _, err := l.Insert(42, 9); err == nil || !strings.Contains(err.Error(), "42") {
		t.Errorf("Insert(42) error = %v, want it to mention index 42", err)
	}
	if _, err := l.Insert(-1, 9); err == nil || !strings.Contains(err.Error(), "-1") {
		t.Errorf("Insert(-1) error = %v, want it to mention index -1", err)
	}
}

func TestRemove(t *testing.T) {
	l := collection.Of(1, 2, 3, 2)

	if got := collection.Remove(l, 2).ToSlice(); !slices.Equal(got, []int{1, 3}) {
		t.Errorf("Remove = %v, want [1 3]", got)
	}
	if got := collection.Remove(l, 9).ToSlice(); !slices.Equal(got, []int{1, 2, 3, 2}) {
		t.Errorf("removing an absent value must change nothing, got %v", got)
	}
}

func TestString(t *testing.T) {
	if got := collection.Of(1, 2, 3).String(); got != "List(1, 2, 3)" {
		t.Errorf("String = %q, want %q", got, "List(1, 2, 3)")
	}
	if got := collection.Empty[int]().String(); got != "List()" {
		t.Errorf("empty String = %q, want %q", got, "List()")
	}
}

// Regression guard: building a list used to be O(n^2) in allocations.
func TestBuildingIsLinear(t *testing.T) {
	alloc := func(n int) float64 {
		values := make([]int, n)
		return testing.AllocsPerRun(1, func() { collection.Of(values...) })
	}

	small, large := alloc(1000), alloc(2000)
	if large > 3*small {
		t.Errorf("building 2000 elements took %.0f allocations vs %.0f for 1000: not linear", large, small)
	}
}

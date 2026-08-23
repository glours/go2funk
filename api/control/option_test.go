package control_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/glours/go2funk/api/control"
)

func TestZeroValueIsNone(t *testing.T) {
	var o control.Option[int]

	if !o.IsEmpty() {
		t.Error("the zero value of Option must be None")
	}
	if got := o.OrElse(7); got != 7 {
		t.Errorf("OrElse = %d, want 7", got)
	}
}

func TestSomeAndNone(t *testing.T) {
	some := control.Some(10)
	none := control.None[int]()

	if !some.IsDefined() || some.IsEmpty() {
		t.Error("Some must be defined")
	}
	if none.IsDefined() || !none.IsEmpty() {
		t.Error("None must be empty")
	}
}

func TestGet(t *testing.T) {
	if value, ok := control.Some(10).Get(); !ok || value != 10 {
		t.Errorf("Get = (%d, %v), want (10, true)", value, ok)
	}
	if value, ok := control.None[int]().Get(); ok || value != 0 {
		t.Errorf("Get = (%d, %v), want (0, false)", value, ok)
	}
}

func TestOrElse(t *testing.T) {
	if got := control.Some(10).OrElse(5); got != 10 {
		t.Errorf("Some.OrElse = %d, want 10", got)
	}
	if got := control.None[int]().OrElse(5); got != 5 {
		t.Errorf("None.OrElse = %d, want 5", got)
	}
}

func TestOrElseGet(t *testing.T) {
	called := false
	fallback := func() int { called = true; return 5 }

	if got := control.Some(10).OrElseGet(fallback); got != 10 {
		t.Errorf("Some.OrElseGet = %d, want 10", got)
	}
	if called {
		t.Error("the fallback must not be evaluated when a value is present")
	}
	if got := control.None[int]().OrElseGet(fallback); got != 5 {
		t.Errorf("None.OrElseGet = %d, want 5", got)
	}
}

func TestOrElseError(t *testing.T) {
	missing := errors.New("no value")

	value, err := control.Some(10).OrElseError(missing)
	if err != nil || value != 10 {
		t.Errorf("Some.OrElseError = (%d, %v), want (10, nil)", value, err)
	}

	_, err = control.None[int]().OrElseError(missing)
	if !errors.Is(err, missing) {
		t.Errorf("None.OrElseError = %v, want %v", err, missing)
	}
}

// The whole point of the rewrite: Map changes the type parameter, as a method.
func TestMapChangesTheType(t *testing.T) {
	var result control.Option[string] = control.Some(10).Map(strconv.Itoa)

	if got := result.OrElse("none"); got != "10" {
		t.Errorf("Map = %q, want %q", got, "10")
	}
	if got := control.None[int]().Map(strconv.Itoa).OrElse("none"); got != "none" {
		t.Errorf("None.Map = %q, want %q", got, "none")
	}
}

func TestMapIsNotEvaluatedOnNone(t *testing.T) {
	control.None[int]().Map(func(int) string {
		t.Error("the mapper must not run on a None")
		return ""
	})
}

func TestChaining(t *testing.T) {
	isEven := func(value int) bool { return value%2 == 0 }

	got := control.Some(10).
		Map(strconv.Itoa).
		Filter(func(s string) bool { return s != "" }).
		Map(func(s string) int { return len(s) }).
		Filter(isEven).
		OrElse(-1)

	if got != 2 {
		t.Errorf("chain = %d, want 2", got)
	}
}

func TestFlatMap(t *testing.T) {
	parse := func(s string) control.Option[int] {
		value, err := strconv.Atoi(s)
		if err != nil {
			return control.None[int]()
		}
		return control.Some(value)
	}

	if got := control.Some("42").FlatMap(parse).OrElse(-1); got != 42 {
		t.Errorf("FlatMap = %d, want 42", got)
	}
	if got := control.Some("nope").FlatMap(parse).OrElse(-1); got != -1 {
		t.Errorf("FlatMap on a failing parse = %d, want -1", got)
	}
	if got := control.None[string]().FlatMap(parse).OrElse(-1); got != -1 {
		t.Errorf("None.FlatMap = %d, want -1", got)
	}
}

func TestFilter(t *testing.T) {
	isEven := func(value int) bool { return value%2 == 0 }

	if control.Some(10).Filter(isEven).IsEmpty() {
		t.Error("Filter must keep a matching value")
	}
	if !control.Some(11).Filter(isEven).IsEmpty() {
		t.Error("Filter must drop a value that fails the predicate")
	}
	if !control.None[int]().Filter(isEven).IsEmpty() {
		t.Error("Filter on None stays None")
	}
}

func TestFold(t *testing.T) {
	onNone := func() string { return "empty" }
	onSome := strconv.Itoa

	if got := control.Some(10).Fold(onNone, onSome); got != "10" {
		t.Errorf("Some.Fold = %q, want %q", got, "10")
	}
	if got := control.None[int]().Fold(onNone, onSome); got != "empty" {
		t.Errorf("None.Fold = %q, want %q", got, "empty")
	}
}

func TestOr(t *testing.T) {
	if got := control.Some(10).Or(control.Some(5)).OrElse(-1); got != 10 {
		t.Errorf("Some.Or = %d, want 10", got)
	}
	if got := control.None[int]().Or(control.Some(5)).OrElse(-1); got != 5 {
		t.Errorf("None.Or = %d, want 5", got)
	}
}

func TestForEach(t *testing.T) {
	seen := 0
	control.Some(10).ForEach(func(value int) { seen = value })
	if seen != 10 {
		t.Errorf("ForEach saw %d, want 10", seen)
	}

	control.None[int]().ForEach(func(int) { t.Error("ForEach must not run on a None") })
}

func TestGoInterop(t *testing.T) {
	value := 10

	if got := control.FromPointer(&value).OrElse(-1); got != 10 {
		t.Errorf("FromPointer = %d, want 10", got)
	}
	if !control.FromPointer[int](nil).IsEmpty() {
		t.Error("FromPointer(nil) must be None")
	}
	if got := control.FromTuple(10, true).OrElse(-1); got != 10 {
		t.Errorf("FromTuple = %d, want 10", got)
	}
	if !control.FromTuple(10, false).IsEmpty() {
		t.Error("FromTuple with ok=false must be None")
	}

	if p := control.Some(10).ToPointer(); p == nil || *p != 10 {
		t.Error("ToPointer must return a pointer to the value")
	}
	if control.None[int]().ToPointer() != nil {
		t.Error("None.ToPointer must be nil")
	}
	if got := control.Some(10).ToSlice(); len(got) != 1 || got[0] != 10 {
		t.Errorf("ToSlice = %v, want [10]", got)
	}
	if got := control.None[int]().ToSlice(); len(got) != 0 {
		t.Errorf("None.ToSlice = %v, want []", got)
	}
}

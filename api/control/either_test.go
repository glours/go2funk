package control_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/glours/go2funk/api/control"
)

var errBoom = errors.New("boom")

func TestEitherSides(t *testing.T) {
	right := control.Right[string](10)
	left := control.Left[string, int]("nope")

	if !right.IsRight() || right.IsLeft() {
		t.Error("Right must be right")
	}
	if left.IsRight() || !left.IsLeft() {
		t.Error("Left must be left")
	}
}

func TestEitherZeroValueIsLeft(t *testing.T) {
	var e control.Either[string, int]

	if !e.IsLeft() {
		t.Error("the zero value of Either must be a Left")
	}
	if got := e.OrElse(7); got != 7 {
		t.Errorf("OrElse = %d, want 7", got)
	}
}

func TestEitherGet(t *testing.T) {
	right := control.Right[string](10)
	left := control.Left[string, int]("nope")

	if value, ok := right.Get(); !ok || value != 10 {
		t.Errorf("Right.Get = (%d, %v), want (10, true)", value, ok)
	}
	if _, ok := left.Get(); ok {
		t.Error("Left.Get must report false")
	}
	if value, ok := left.GetLeft(); !ok || value != "nope" {
		t.Errorf("Left.GetLeft = (%q, %v), want (\"nope\", true)", value, ok)
	}
	if _, ok := right.GetLeft(); ok {
		t.Error("Right.GetLeft must report false")
	}
}

func TestEitherMapChangesTheRightType(t *testing.T) {
	var result control.Either[string, string] = control.Right[string](10).Map(strconv.Itoa)

	if got := result.OrElse("none"); got != "10" {
		t.Errorf("Map = %q, want %q", got, "10")
	}
}

// A Left must survive Map with its value intact — the bug that used to bite Try.
func TestEitherMapKeepsTheLeft(t *testing.T) {
	left := control.Left[string, int]("nope")

	mapped := left.Map(strconv.Itoa)
	if !mapped.IsLeft() {
		t.Fatal("mapping a Left must stay a Left")
	}
	if value, _ := mapped.GetLeft(); value != "nope" {
		t.Errorf("left value = %q, want %q", value, "nope")
	}
	left.Map(func(int) string {
		t.Error("the mapper must not run on a Left")
		return ""
	})
}

func TestEitherFlatMapAndMapLeft(t *testing.T) {
	double := func(value int) control.Either[string, int] { return control.Right[string](value * 2) }

	if got := control.Right[string](10).FlatMap(double).OrElse(-1); got != 20 {
		t.Errorf("FlatMap = %d, want 20", got)
	}
	if got := control.Left[string, int]("nope").FlatMap(double).OrElse(-1); got != -1 {
		t.Errorf("Left.FlatMap = %d, want -1", got)
	}

	upper := control.Left[string, int]("nope").MapLeft(func(s string) int { return len(s) })
	if value, _ := upper.GetLeft(); value != 4 {
		t.Errorf("MapLeft = %d, want 4", value)
	}
}

func TestEitherSwapFoldFilter(t *testing.T) {
	right := control.Right[string](10)

	if value, ok := right.Swap().GetLeft(); !ok || value != 10 {
		t.Errorf("Swap = (%d, %v), want (10, true)", value, ok)
	}

	fold := func(e control.Either[string, int]) string {
		return e.Fold(func(s string) string { return "left:" + s }, strconv.Itoa)
	}
	if got := fold(right); got != "10" {
		t.Errorf("Fold on Right = %q, want %q", got, "10")
	}
	if got := fold(control.Left[string, int]("nope")); got != "left:nope" {
		t.Errorf("Fold on Left = %q, want %q", got, "left:nope")
	}

	isEven := func(value int) bool { return value%2 == 0 }
	tooOdd := func(value int) string { return "odd" }
	if !right.FilterOrElse(isEven, tooOdd).IsRight() {
		t.Error("FilterOrElse must keep a matching Right")
	}
	if value, _ := control.Right[string](11).FilterOrElse(isEven, tooOdd).GetLeft(); value != "odd" {
		t.Errorf("FilterOrElse = %q, want %q", value, "odd")
	}
}

func TestEitherToOption(t *testing.T) {
	if got := control.Right[string](10).ToOption().OrElse(-1); got != 10 {
		t.Errorf("Right.ToOption = %d, want 10", got)
	}
	if !control.Left[string, int]("nope").ToOption().IsEmpty() {
		t.Error("Left.ToOption must be None")
	}
}

// Result is an alias of Either[error, T]: it inherits every method above.
func TestResultIsAnEitherAlias(t *testing.T) {
	var _ control.Result[int] = control.Right[error](10)
	var _ control.Either[error, int] = control.Ok(10)

	if got := control.Ok(21).Map(func(v int) int { return v * 2 }).Map(strconv.Itoa).OrElse("?"); got != "42" {
		t.Errorf("chained Result = %q, want %q", got, "42")
	}
}

func TestTryCapturesTheError(t *testing.T) {
	ok := control.Try(func() (int, error) { return 10, nil })
	failed := control.Try(func() (int, error) { return 0, errBoom })

	if !ok.IsRight() {
		t.Error("Try must return an Ok when the function succeeds")
	}
	if !failed.IsLeft() {
		t.Fatal("Try must return an Err when the function fails")
	}

	// The regression that mattered: the cause must survive.
	if _, err := control.Unwrap(failed); !errors.Is(err, errBoom) {
		t.Errorf("cause = %v, want %v", err, errBoom)
	}
	if _, err := control.Unwrap(failed.Map(strconv.Itoa)); !errors.Is(err, errBoom) {
		t.Errorf("cause after Map = %v, want %v", err, errBoom)
	}
}

func TestUnwrap(t *testing.T) {
	value, err := control.Unwrap(control.Ok(10))
	if err != nil || value != 10 {
		t.Errorf("Unwrap(Ok) = (%d, %v), want (10, nil)", value, err)
	}

	value, err = control.Unwrap(control.Err[int](errBoom))
	if !errors.Is(err, errBoom) || value != 0 {
		t.Errorf("Unwrap(Err) = (%d, %v), want (0, boom)", value, err)
	}
}

func TestEitherOrElseGet(t *testing.T) {
	called := false
	fallback := func(s string) int { called = true; return len(s) }

	if got := control.Right[string](10).OrElseGet(fallback); got != 10 {
		t.Errorf("Right.OrElseGet = %d, want 10", got)
	}
	if called {
		t.Error("the fallback must not be evaluated on a Right")
	}
	if got := control.Left[string, int]("nope").OrElseGet(fallback); got != 4 {
		t.Errorf("Left.OrElseGet = %d, want 4", got)
	}
}

func TestEitherLeftOrElseAndOr(t *testing.T) {
	right := control.Right[string](10)
	left := control.Left[string, int]("nope")

	if got := right.LeftOrElse("fallback"); got != "fallback" {
		t.Errorf("Right.LeftOrElse = %q, want %q", got, "fallback")
	}
	if got := left.LeftOrElse("fallback"); got != "nope" {
		t.Errorf("Left.LeftOrElse = %q, want %q", got, "nope")
	}

	if got := right.Or(control.Right[string](5)).OrElse(-1); got != 10 {
		t.Errorf("Right.Or = %d, want 10", got)
	}
	if got := left.Or(control.Right[string](5)).OrElse(-1); got != 5 {
		t.Errorf("Left.Or = %d, want 5", got)
	}
}

func TestEitherSwapOnLeftAndMapLeftOnRight(t *testing.T) {
	swapped := control.Left[string, int]("nope").Swap()
	if value, ok := swapped.Get(); !ok || value != "nope" {
		t.Errorf("Left.Swap = (%q, %v), want (\"nope\", true)", value, ok)
	}

	mapped := control.Right[string](10).MapLeft(func(s string) int { return len(s) })
	if value, ok := mapped.Get(); !ok || value != 10 {
		t.Errorf("Right.MapLeft must keep the right value, got (%d, %v)", value, ok)
	}
}

func TestEitherForEach(t *testing.T) {
	seen := 0
	control.Right[string](10).ForEach(func(value int) { seen = value })
	if seen != 10 {
		t.Errorf("ForEach saw %d, want 10", seen)
	}

	control.Left[string, int]("nope").ForEach(func(int) { t.Error("ForEach must not run on a Left") })
}

func TestFilterOrElseOnLeft(t *testing.T) {
	left := control.Left[string, int]("nope")

	result := left.FilterOrElse(
		func(int) bool { t.Error("the predicate must not run on a Left"); return false },
		func(int) string { t.Error("orElse must not run on a Left"); return "" },
	)
	if value, _ := result.GetLeft(); value != "nope" {
		t.Errorf("Left.FilterOrElse = %q, want %q", value, "nope")
	}
}

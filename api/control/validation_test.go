package control_test

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/glours/go2funk/api/control"
	"github.com/glours/go2funk/api/tuple"
)

// No error collected means nothing was found wrong: the zero value is valid.
func TestValidationZeroValueIsValid(t *testing.T) {
	var v control.Validation[string, int]

	if !v.IsValid() || v.IsInvalid() {
		t.Error("the zero value of Validation must be valid")
	}
	if value, ok := v.Get(); !ok || value != 0 {
		t.Errorf("Get = (%d, %t), want (0, true)", value, ok)
	}
	if got := v.Errors(); len(got) != 0 {
		t.Errorf("Errors = %v, want none", got)
	}
}

func TestValidationValidAndInvalid(t *testing.T) {
	tests := []struct {
		name       string
		validation control.Validation[string, int]
		valid      bool
		value      int
		errors     []string
	}{
		{name: "valid", validation: control.Valid[string](10), valid: true, value: 10, errors: nil},
		{name: "one error", validation: control.Invalid[string, int]("too small"), errors: []string{"too small"}},
		{name: "several errors, in order", validation: control.Invalid[string, int]("a", "b", "c"), errors: []string{"a", "b", "c"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v := test.validation
			if v.IsValid() != test.valid || v.IsInvalid() == test.valid {
				t.Errorf("IsValid = %t, IsInvalid = %t, want valid %t", v.IsValid(), v.IsInvalid(), test.valid)
			}
			if value, ok := v.Get(); ok != test.valid || value != test.value {
				t.Errorf("Get = (%d, %t), want (%d, %t)", value, ok, test.value, test.valid)
			}
			if got := v.Errors(); !slices.Equal(got, test.errors) {
				t.Errorf("Errors = %v, want %v", got, test.errors)
			}
			if got := v.OrElse(-1); test.valid && got != test.value || !test.valid && got != -1 {
				t.Errorf("OrElse(-1) = %d", got)
			}
		})
	}
}

// A Validation is immutable: neither the slice handed to Invalid nor the one
// returned by Errors can reach back into it.
func TestValidationErrorsCannotBeMutatedFromOutside(t *testing.T) {
	more := []string{"b", "c"}
	v := control.Invalid[string, int]("a", more...)

	more[0] = "changed"
	if got := v.Errors(); !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Fatalf("mutating the slice given to Invalid changed it: %v", got)
	}

	v.Errors()[0] = "changed"
	if got := v.Errors(); !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Errorf("mutating the slice returned by Errors changed it: %v", got)
	}
}

func TestValidationMapChangesTheValueType(t *testing.T) {
	var mapped control.Validation[string, string] = control.Valid[string](10).Map(strconv.Itoa)
	if value, ok := mapped.Get(); !ok || value != "10" {
		t.Errorf("Map = (%q, %t), want (\"10\", true)", value, ok)
	}

	called := false
	invalid := control.Invalid[string, int]("a", "b").Map(func(int) string { called = true; return "" })
	if called {
		t.Error("Map must not call the mapper on an invalid validation")
	}
	if got := invalid.Errors(); !slices.Equal(got, []string{"a", "b"}) {
		t.Errorf("Map dropped the errors: %v, want [a b]", got)
	}
}

func TestValidationMapErrorMapsEveryError(t *testing.T) {
	var mapped control.Validation[int, string] = control.Invalid[string, string]("a", "bb", "ccc").
		MapError(func(err string) int { return len(err) })
	if got := mapped.Errors(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("MapError = %v, want [1 2 3]", got)
	}

	called := false
	valid := control.Valid[string]("ok").MapError(func(string) int { called = true; return 0 })
	if called {
		t.Error("MapError must not call the mapper on a valid validation")
	}
	if value, ok := valid.Get(); !ok || value != "ok" {
		t.Errorf("MapError on a valid validation = (%q, %t), want (\"ok\", true)", value, ok)
	}
}

// FlatMap is sequential: the next step needs the value, so an invalid
// validation stops there and keeps its own errors only.
func TestValidationFlatMapStopsAtTheFirstInvalid(t *testing.T) {
	positive := func(value int) control.Validation[string, int] {
		if value <= 0 {
			return control.Invalid[string, int]("not positive")
		}
		return control.Valid[string](value)
	}

	tests := []struct {
		name   string
		input  control.Validation[string, int]
		value  int
		errors []string
	}{
		{name: "valid into valid", input: control.Valid[string](3), value: 3},
		{name: "valid into invalid", input: control.Valid[string](-3), errors: []string{"not positive"}},
		{name: "invalid skips the step", input: control.Invalid[string, int]("earlier"), errors: []string{"earlier"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.input.FlatMap(positive)
			if value, _ := got.Get(); value != test.value {
				t.Errorf("value = %d, want %d", value, test.value)
			}
			if errs := got.Errors(); !slices.Equal(errs, test.errors) {
				t.Errorf("Errors = %v, want %v", errs, test.errors)
			}
		})
	}
}

func TestValidationFold(t *testing.T) {
	describe := func(v control.Validation[string, int]) string {
		return v.Fold(
			func(errs []string) string { return "invalid: " + strconv.Itoa(len(errs)) },
			func(value int) string { return "valid: " + strconv.Itoa(value) },
		)
	}

	if got := describe(control.Valid[string](7)); got != "valid: 7" {
		t.Errorf("Fold on valid = %q", got)
	}
	if got := describe(control.Invalid[string, int]("a", "b")); got != "invalid: 2" {
		t.Errorf("Fold on invalid = %q", got)
	}

	// onInvalid gets its own copy, like Errors.
	v := control.Invalid[string, int]("a")
	v.Fold(func(errs []string) int { errs[0] = "changed"; return 0 }, func(int) int { return 0 })
	if got := v.Errors(); !slices.Equal(got, []string{"a"}) {
		t.Errorf("Fold let onInvalid change the validation: %v", got)
	}
}

// form is the fixture for the accumulating tests: three independent fields.
type form struct {
	name  string
	email string
	age   int
}

func checkName(f form) control.Validation[string, form] {
	if f.name == "" {
		return control.Invalid[string, form]("name is empty")
	}
	return control.Valid[string](f)
}

func checkEmail(f form) control.Validation[string, form] {
	if !strings.Contains(f.email, "@") {
		return control.Invalid[string, form]("email has no @")
	}
	return control.Valid[string](f)
}

func checkAge(f form) control.Validation[string, form] {
	if f.age <= 0 {
		return control.Invalid[string, form]("age is not positive", "age is required")
	}
	return control.Valid[string](f)
}

func TestCheckRunsEveryCheck(t *testing.T) {
	tests := []struct {
		name   string
		input  form
		errors []string
	}{
		{name: "all pass", input: form{name: "ada", email: "ada@example.com", age: 36}},
		{name: "one fails", input: form{name: "", email: "ada@example.com", age: 36}, errors: []string{"name is empty"}},
		{
			name:   "all fail, errors in the order of the checks",
			input:  form{},
			errors: []string{"name is empty", "email has no @", "age is not positive", "age is required"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := control.Check(test.input, checkName, checkEmail, checkAge)
			if errs := got.Errors(); !slices.Equal(errs, test.errors) {
				t.Errorf("Errors = %v, want %v", errs, test.errors)
			}
			if value, ok := got.Get(); ok != (test.errors == nil) || ok && value != test.input {
				t.Errorf("Get = (%v, %t), want (%v, %t)", value, ok, test.input, test.errors == nil)
			}
		})
	}

	if value, ok := control.Check[string](form{name: "x"}).Get(); !ok || value.name != "x" {
		t.Errorf("Check with no check = (%v, %t), want the value, valid", value, ok)
	}
}

func TestSequence(t *testing.T) {
	tests := []struct {
		name        string
		validations []control.Validation[string, int]
		values      []int
		errors      []string
	}{
		{name: "none", validations: nil, values: []int{}},
		{
			name:        "all valid, values in order",
			validations: []control.Validation[string, int]{control.Valid[string](1), control.Valid[string](2), control.Valid[string](3)},
			values:      []int{1, 2, 3},
		},
		{
			name: "invalid ones, errors in order",
			validations: []control.Validation[string, int]{
				control.Invalid[string, int]("a"), control.Valid[string](2), control.Invalid[string, int]("b", "c"),
			},
			errors: []string{"a", "b", "c"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := control.Sequence(test.validations...)
			if errs := got.Errors(); !slices.Equal(errs, test.errors) {
				t.Errorf("Errors = %v, want %v", errs, test.errors)
			}
			if values, _ := got.Get(); !slices.Equal(values, test.values) {
				t.Errorf("values = %v, want %v", values, test.values)
			}
		})
	}
}

func TestZip(t *testing.T) {
	name := control.Valid[string]("ada")
	age := control.Valid[string](36)

	var zipped control.Validation[string, tuple.Pair[string, int]] = control.Zip(name, age)
	if pair, ok := zipped.Get(); !ok || pair != tuple.New("ada", 36) {
		t.Errorf("Zip of two valid = (%v, %t), want ({ada 36}, true)", pair, ok)
	}

	tests := []struct {
		name   string
		left   control.Validation[string, string]
		right  control.Validation[string, int]
		errors []string
	}{
		{name: "left invalid", left: control.Invalid[string, string]("a"), right: age, errors: []string{"a"}},
		{name: "right invalid", left: name, right: control.Invalid[string, int]("b"), errors: []string{"b"}},
		{
			name:   "both invalid, left errors first",
			left:   control.Invalid[string, string]("a1", "a2"),
			right:  control.Invalid[string, int]("b"),
			errors: []string{"a1", "a2", "b"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if errs := control.Zip(test.left, test.right).Errors(); !slices.Equal(errs, test.errors) {
				t.Errorf("Errors = %v, want %v", errs, test.errors)
			}
		})
	}
}

// Zip nests without an arity limit, at the cost of nesting the pairs.
func TestZipChains(t *testing.T) {
	zipped := control.Zip(control.Zip(control.Valid[string]("ada"), control.Valid[string]("ada@example.com")), control.Valid[string](36))

	var nested tuple.Pair[tuple.Pair[string, string], int]
	nested, ok := zipped.Get()
	if !ok || nested != tuple.New(tuple.New("ada", "ada@example.com"), 36) {
		t.Errorf("Zip chain = (%v, %t)", nested, ok)
	}

	failed := control.Zip(control.Zip(control.Invalid[string, string]("a"), control.Valid[string](1)), control.Invalid[string, int]("c"))
	if errs := failed.Errors(); !slices.Equal(errs, []string{"a", "c"}) {
		t.Errorf("Errors = %v, want [a c]", errs)
	}
}

// Accumulating twice from the same invalid validation must give two
// independent results. Appending to its errors in place would let the second
// accumulation overwrite the first whenever the slice has spare capacity — and
// a validation that is itself the result of an accumulation usually has some,
// since append grows the slice ahead of need: three errors gathered one at a
// time sit in a slice with room for four.
func TestAccumulatingDoesNotShareErrors(t *testing.T) {
	base := control.Sequence(
		control.Invalid[string, int]("a"),
		control.Invalid[string, int]("b"),
		control.Invalid[string, int]("c"),
	).Map(func([]int) int { return 0 })

	tests := []struct {
		name       string
		accumulate func(other string) control.Validation[string, int]
	}{
		{name: "Zip", accumulate: func(other string) control.Validation[string, int] {
			return control.Zip(base, control.Invalid[string, int](other)).Map(func(tuple.Pair[int, int]) int { return 0 })
		}},
		{name: "Sequence", accumulate: func(other string) control.Validation[string, int] {
			return control.Sequence(base, control.Invalid[string, int](other)).Map(func([]int) int { return 0 })
		}},
		{name: "Check", accumulate: func(other string) control.Validation[string, int] {
			return control.Check(0,
				func(int) control.Validation[string, int] { return base },
				func(int) control.Validation[string, int] { return control.Invalid[string, int](other) })
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			first := test.accumulate("x")
			second := test.accumulate("y")
			if errs := first.Errors(); !slices.Equal(errs, []string{"a", "b", "c", "x"}) {
				t.Errorf("first = %v, want [a b c x]", errs)
			}
			if errs := second.Errors(); !slices.Equal(errs, []string{"a", "b", "c", "y"}) {
				t.Errorf("second = %v, want [a b c y]", errs)
			}
			if errs := base.Errors(); !slices.Equal(errs, []string{"a", "b", "c"}) {
				t.Errorf("base = %v, want [a b c]", errs)
			}
		})
	}
}

func TestFromEither(t *testing.T) {
	if value, ok := control.FromEither(control.Right[string](3)).Get(); !ok || value != 3 {
		t.Errorf("FromEither(Right(3)) = (%d, %t), want (3, true)", value, ok)
	}
	if errs := control.FromEither(control.Left[string, int]("nope")).Errors(); !slices.Equal(errs, []string{"nope"}) {
		t.Errorf("FromEither(Left) errors = %v, want [nope]", errs)
	}

	// Result is an Either, so it converts the same way, keeping its cause.
	if errs := control.FromEither(control.Err[int](errBoom)).Errors(); len(errs) != 1 || !errors.Is(errs[0], errBoom) {
		t.Errorf("FromEither(Err(boom)) errors = %v, want [boom]", errs)
	}
}

func TestValidationToEither(t *testing.T) {
	var right control.Either[[]string, int] = control.Valid[string](3).ToEither()
	if value, ok := right.Get(); !ok || value != 3 {
		t.Errorf("ToEither of valid = (%d, %t), want (3, true)", value, ok)
	}

	v := control.Invalid[string, int]("a", "b")
	left, ok := v.ToEither().GetLeft()
	if !ok || !slices.Equal(left, []string{"a", "b"}) {
		t.Fatalf("ToEither of invalid = (%v, %t), want ([a b], true)", left, ok)
	}
	left[0] = "changed"
	if errs := v.Errors(); !slices.Equal(errs, []string{"a", "b"}) {
		t.Errorf("mutating the Either's errors changed the validation: %v", errs)
	}
}

func TestValidationToOption(t *testing.T) {
	if value, ok := control.Valid[string](3).ToOption().Get(); !ok || value != 3 {
		t.Errorf("ToOption of valid = (%d, %t), want (3, true)", value, ok)
	}
	if !control.Invalid[string, int]("a").ToOption().IsEmpty() {
		t.Error("ToOption of invalid must be None")
	}
}

// Every cause survives the conversion and stays reachable through errors.Is,
// which is what joining them buys over keeping only the first.
func TestToResultJoinsEveryCause(t *testing.T) {
	errFirst, errSecond := errors.New("first"), errors.New("second")

	value, err := control.Unwrap(control.ToResult(control.Valid[error](3)))
	if err != nil || value != 3 {
		t.Errorf("ToResult of valid = (%d, %v), want (3, nil)", value, err)
	}

	_, err = control.Unwrap(control.ToResult(control.Invalid[error, int](errFirst, errSecond)))
	if !errors.Is(err, errFirst) || !errors.Is(err, errSecond) {
		t.Fatalf("ToResult lost a cause: %v", err)
	}
	if got := err.Error(); got != "first\nsecond" {
		t.Errorf("Error() = %q, want the causes one per line, as errors.Join renders them", got)
	}
}

// errors.Join drops nil errors, so an invalid validation holding only nil ones
// joins to nil. The Result is still a failure, as Err(nil) already is.
func TestToResultOfNilErrorsIsStillAFailure(t *testing.T) {
	result := control.ToResult(control.Invalid[error, int](nil))

	if !result.IsLeft() {
		t.Error("an invalid validation must convert to a failed Result")
	}
	if _, err := control.Unwrap(result); err != nil {
		t.Errorf("Unwrap error = %v, want nil: errors.Join drops nil errors", err)
	}
}

// Invalid knows how many errors it holds, so it builds their slice once at the
// right size rather than growing it.
func TestInvalidAllocatesOnce(t *testing.T) {
	var kept control.Validation[string, int]
	tests := []struct {
		name string
		more []string
	}{
		{name: "one error", more: nil},
		{name: "three errors", more: []string{"b", "c"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			allocations := testing.AllocsPerRun(50, func() { kept = control.Invalid[string, int]("a", test.more...) })
			if allocations != 1 {
				t.Errorf("Invalid allocated %.0f times, want 1", allocations)
			}
		})
	}
	_ = kept
}

// Sequence counts the errors before building anything, so it knows up front
// whether it is gathering values or errors, and how many: one allocation, of
// the exact size, whichever way it goes. Building the values slice and growing
// the errors one would waste both.
func TestSequenceAllocatesOnce(t *testing.T) {
	valid := make([]control.Validation[string, int], 1000)
	invalid := make([]control.Validation[string, int], 1000)
	for index := range 1000 {
		valid[index] = control.Valid[string](index)
		if index%2 == 0 {
			invalid[index] = control.Invalid[string, int]("even")
		} else {
			invalid[index] = control.Valid[string](index)
		}
	}

	tests := []struct {
		name        string
		validations []control.Validation[string, int]
	}{
		{name: "all valid", validations: valid},
		{name: "half invalid", validations: invalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var kept control.Validation[string, []int]
			allocations := testing.AllocsPerRun(20, func() { kept = control.Sequence(test.validations...) })
			if allocations != 1 {
				t.Errorf("Sequence of 1000 allocated %.0f times, want 1", allocations)
			}
			_ = kept
		})
	}
}

// Sequence costs one pass to count and one to gather, with a single
// allocation of the exact size either way.
func BenchmarkSequence(b *testing.B) {
	for _, size := range []int{10, 1000} {
		valid := make([]control.Validation[string, int], size)
		invalid := make([]control.Validation[string, int], size)
		for index := range size {
			valid[index] = control.Valid[string](index)
			invalid[index] = control.Invalid[string, int]("invalid")
		}
		b.Run(fmt.Sprintf("valid/%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = control.Sequence(valid...)
			}
		})
		b.Run(fmt.Sprintf("invalid/%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = control.Sequence(invalid...)
			}
		})
	}
}

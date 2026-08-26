package control_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/glours/go2funk/api/control"
)

func ExampleOption() {
	some := control.Some(10)
	none := control.None[int]()

	fmt.Println(some.OrElse(5), none.OrElse(5))

	// Map is a method and changes the type it carries.
	fmt.Println(some.Map(strconv.Itoa).Map(strings.ToUpper).OrElse("none"))

	isEven := func(value int) bool { return value%2 == 0 }
	fmt.Println(some.Filter(isEven).IsDefined())

	// The zero value is None, not a nil panic.
	var zero control.Option[int]
	fmt.Println(zero.IsEmpty())

	// Output:
	// 10 5
	// 10
	// true
	// true
}

func ExampleOption_goInterop() {
	counts := map[string]int{"ten": 10}

	// The Go "comma ok" idiom lifts straight into an Option.
	value, ok := counts["ten"]
	found := control.FromTuple(value, ok)

	value, ok = counts["nope"]
	missing := control.FromTuple(value, ok)

	fmt.Println(found.OrElse(-1), missing.OrElse(-1))

	value, err := missing.OrElseError(errors.New("no value"))
	fmt.Println(value, err)

	fmt.Println(found.ToSlice(), control.None[int]().ToPointer() == nil)

	// Output:
	// 10 -1
	// 0 no value
	// [10] true
}

func ExampleEither() {
	right := control.Right[string](10)
	left := control.Left[string, int]("nope")

	fmt.Println(right.OrElse(20), left.OrElse(20))

	// Map works on the right side; a Left passes through with its value intact.
	fmt.Println(right.Map(strconv.Itoa).OrElse("none"))
	fmt.Println(left.Map(strconv.Itoa).LeftOrElse("?"))

	fmt.Println(right.Fold(
		func(s string) string { return "left: " + s },
		func(v int) string { return "right: " + strconv.Itoa(v) },
	))

	// Output:
	// 10 20
	// 10
	// nope
	// right: 10
}

func ExampleResult() {
	parse := func(s string) control.Result[int] {
		return control.Try(func() (int, error) { return strconv.Atoi(s) })
	}

	fmt.Println(parse("42").Map(func(v int) int { return v * 2 }).OrElse(-1))

	// The cause survives Map, and Unwrap hands it back to the Go idiom.
	_, err := control.Unwrap(parse("nope").Map(strconv.Itoa))
	fmt.Println(err)

	fmt.Println(parse("nope").ToOption().IsEmpty())

	// Output:
	// 84
	// strconv.Atoi: parsing "nope": invalid syntax
	// true
}

func ExampleLazy() {
	calls := 0
	expensive := control.NewLazy(func() int {
		calls++
		fmt.Println("computing...")
		return 42
	})

	fmt.Println("created, calls:", calls)

	fmt.Println(expensive.Get())
	fmt.Println(expensive.Get())
	fmt.Println("calls:", calls, "evaluated:", expensive.IsEvaluated())

	// Output:
	// created, calls: 0
	// computing...
	// 42
	// 42
	// calls: 1 evaluated: true
}

func ExampleLazy_Map() {
	source := control.NewLazy(func() int {
		fmt.Println("source runs")
		return 21
	})

	// Map stays lazy: nothing runs here.
	doubled := source.Map(func(value int) int { return value * 2 }).Map(strconv.Itoa)
	fmt.Println("mapped, evaluated:", doubled.IsEvaluated())

	fmt.Println(doubled.Get())

	// The zero value has no computation and yields the zero value of T.
	var empty control.Lazy[int]
	fmt.Println(empty.Get(), empty.IsEvaluated())

	// Output:
	// mapped, evaluated: false
	// source runs
	// 42
	// 0 true
}

func ExampleOption_json() {
	type Profile struct {
		Name     string                 `json:"name"`
		Nickname control.Option[string] `json:"nickname"`
		Age      control.Option[int]    `json:"age,omitzero"`
	}

	encoded, _ := json.Marshal(Profile{Name: "ada", Nickname: control.Some("countess")})
	fmt.Println(string(encoded))

	var back Profile
	_ = json.Unmarshal([]byte(`{"name":"ada","nickname":null,"age":36}`), &back)
	fmt.Println(back.Nickname.OrElse("none"), back.Age.OrElse(-1))

	// A missing key leaves the Option empty.
	var partial Profile
	_ = json.Unmarshal([]byte(`{"name":"ada"}`), &partial)
	fmt.Println(partial.Nickname.IsEmpty(), partial.Age.IsEmpty())

	// Output:
	// {"name":"ada","nickname":"countess"}
	// none 36
	// true true
}

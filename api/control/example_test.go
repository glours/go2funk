package control_test

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/glours/go2funk/api/control"
)

func ExampleOption() {
	empty := control.Empty[int]()
	some := control.Of(10)

	fmt.Println(empty.OrElse(5))
	fmt.Println(some.OrElse(5))

	isEven := func(value int) bool { return value%2 == 0 }
	fmt.Println(some.Filter(isEven).IsEmpty())

	asString := control.MapOption(some, strconv.Itoa)
	fmt.Println(asString.OrElse("none"))

	// Output:
	// 5
	// 10
	// false
	// 10
}

func ExampleOption_orElseError() {
	missing := errors.New("no value")

	_, err := control.Empty[int]().OrElseError(missing)
	fmt.Println(err)

	value, err := control.Of(10).OrElseError(missing)
	fmt.Println(value, err)

	// Output:
	// no value
	// 10 <nil>
}

func ExampleTry() {
	boom := errors.New("boom")

	success := control.SuccessOf(10)
	failure := control.FailureOf[int](boom)

	fmt.Println(success.IsFailure(), failure.IsFailure())
	fmt.Println(success.OrElse(5), failure.OrElse(5))

	_, err := failure.OrElseCause()
	fmt.Println(err)

	fmt.Println(control.TryOf(func() (int, error) { return 10, nil }).IsFailure())
	fmt.Println(control.TryOf(func() (int, error) { return 0, boom }).IsFailure())

	// Output:
	// false true
	// 10 5
	// boom
	// false
	// true
}

func ExampleEither() {
	boom := errors.New("boom")
	noError := errors.New("no error")

	right := control.RightOf[error](10)
	left := control.LeftOf[error, int](boom)

	fmt.Println(right.IsRight(), left.IsLeft())
	fmt.Println(right.GetOrElse(20), left.GetOrElse(20))
	fmt.Println(right.GetLeftOrElse(noError), left.GetLeftOrElse(noError))

	asString := control.MapEither(right, strconv.Itoa)
	fmt.Println(asString.GetOrElse("none"))

	fmt.Println(right.Swap().GetLeftOrElse(0))

	// Output:
	// true true
	// 10 20
	// no error boom
	// 10
	// 10
}

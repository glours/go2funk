package collection_test

import (
	"fmt"
	"strconv"

	"github.com/glours/go2funk/api/collection"
)

func ExampleList() {
	list := collection.Of(1, 2, 3, 4, 5)

	fmt.Println(list.Length(), list.Head().OrElse(-1))

	isEven := func(value int) bool { return value%2 == 0 }
	fmt.Println(list.Filter(isEven).Map(strconv.Itoa))

	// Prepend is O(1) and the original list is untouched.
	fmt.Println(list.Prepend(0), list)

	// The zero value is the empty list.
	var zero collection.List[int]
	fmt.Println(zero.IsEmpty(), zero.Head().IsEmpty())

	// Output:
	// 5 1
	// List(2, 4)
	// List(0, 1, 2, 3, 4, 5) List(1, 2, 3, 4, 5)
	// true true
}

func ExampleList_All() {
	list := collection.Of("a", "b", "c")

	for value := range list.All() {
		fmt.Println(value)
	}

	fmt.Println(collection.Collect(list.All()).Reverse())

	// Output:
	// a
	// b
	// c
	// List(c, b, a)
}

func ExampleList_Fold() {
	list := collection.Of(1, 2, 3, 4)

	sum := list.Fold(0, func(acc, value int) int { return acc + value })
	joined := list.Fold("", func(acc string, value int) string { return acc + strconv.Itoa(value) })

	fmt.Println(sum, joined)

	// Output:
	// 10 1234
}

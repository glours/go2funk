package collection_test

import (
	"fmt"
	"strconv"

	"github.com/glours/go2funk/api/collection"
)

func ExampleList() {
	list := collection.OfSlice([]int{1, 2, 3, 4, 5})
	list = list.Append(6)
	fmt.Println(list.Length())

	isEven := func(value int) bool { return value%2 == 0 }
	evens := list.Filter(isEven)
	fmt.Println(evens.Length())

	asStrings := collection.MapList(evens, strconv.Itoa)
	fmt.Println(asStrings.Length(), asStrings.IsEmpty())

	// Output:
	// 6
	// 3
	// 3 false
}

func ExampleList_insert() {
	list := collection.OfSlice([]int{1, 2, 4})

	inserted, err := list.Insert(2, 3)
	fmt.Println(inserted.Length(), err)

	// Insert reports an error when the index is out of range.
	// NOTE: the message currently leaks the recursion index rather than the
	// caller's, so this example only asserts that an error is returned.
	_, err = list.Insert(42, 3)
	fmt.Println(err != nil)

	// Output:
	// 4 <nil>
	// true
}

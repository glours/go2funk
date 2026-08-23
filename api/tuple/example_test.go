package tuple_test

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glours/go2funk/api/tuple"
)

func ExamplePair() {
	pair := tuple.New("ten", 10)
	fmt.Println(pair.Left(), pair.Right())

	left, right := pair.Unpack()
	fmt.Println(left, right)

	fmt.Println(pair.MapRight(strconv.Itoa).Right())
	fmt.Println(pair.Map(strings.ToUpper, func(v int) bool { return v > 5 }).Right())
	fmt.Println(pair.Swap().Left())

	// Output:
	// ten 10
	// ten 10
	// 10
	// true
	// 10
}

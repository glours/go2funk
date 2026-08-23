package api_test

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glours/go2funk/api"
)

func ExamplePair() {
	pair := api.NewPair("ten", 10)
	fmt.Println(pair.GetLeft(), pair.GetRight())

	asString := api.MapRightPair(pair, strconv.Itoa)
	fmt.Println(asString.GetRight())

	both := api.MapPair(pair, strings.ToUpper, func(value int) bool { return value > 5 })
	fmt.Println(both.GetLeft(), both.GetRight())

	// Output:
	// ten 10
	// 10
	// TEN true
}

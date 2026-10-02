package collection_test

import (
	"fmt"
	"maps"
	"slices"
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

func ExampleTree() {
	tree := collection.TreeOf(5, 3, 8, 1, 9)

	fmt.Println(tree)
	fmt.Println(tree.Len(), tree.Contains(8), tree.Contains(4))
	fmt.Println(tree.Min().OrElse(-1), tree.Max().OrElse(-1))

	// Subtrees that cannot hold a value in the range are never visited.
	for value := range tree.Range(3, 8) {
		fmt.Println(value)
	}

	// Insert and Delete return new trees; the original is untouched.
	fmt.Println(tree.Insert(4).Delete(9), tree)

	// The zero value is the empty tree.
	var zero collection.Tree[int]
	fmt.Println(zero.IsEmpty(), zero.Min().IsEmpty())

	// Output:
	// Tree(1, 3, 5, 8, 9)
	// 5 true false
	// 1 9
	// 3
	// 5
	// 8
	// Tree(1, 3, 4, 5, 8) Tree(1, 3, 5, 8, 9)
	// true true
}

func ExampleTree_Range() {
	tree := collection.TreeOf(10, 20, 30, 40, 50)

	for value := range tree.Range(20, 40) {
		fmt.Println(value)
	}

	// Iteration is in order, and Backward walks the other way.
	fmt.Println(tree.ToSlice())
	fmt.Println(collection.CollectTree(tree.Backward()).Max().OrElse(-1))

	// Output:
	// 20
	// 30
	// 40
	// [10 20 30 40 50]
	// 50
}

func ExampleMap() {
	ages := collection.EmptyMap[string, int]().Put("ada", 36).Put("alan", 41)

	fmt.Println(ages)
	fmt.Println(ages.Len(), ages.Get("ada").OrElse(-1), ages.Get("grace").IsEmpty())

	// Put and Delete return new maps; the original is untouched.
	fmt.Println(ages.Put("grace", 85).Delete("alan"), ages)

	// Map changes the value type and keeps every key.
	fmt.Println(ages.Map(func(age int) string { return strconv.Itoa(age) + " years" }))

	// The zero value is the empty map.
	var zero collection.Map[string, int]
	fmt.Println(zero.IsEmpty(), zero.Get("ada").IsEmpty())

	// Output:
	// Map[ada:36 alan:41]
	// 2 36 true
	// Map[ada:36 grace:85] Map[ada:36 alan:41]
	// Map[ada:36 years alan:41 years]
	// true true
}

func ExampleCollectMap() {
	// In from a Go map, and back out, through the standard iterators.
	ages := collection.CollectMap(maps.All(map[string]int{"ada": 36, "alan": 41}))
	older := ages.Filter(func(_ string, age int) bool { return age > 40 })

	fmt.Println(maps.Collect(older.All()))

	// Iteration order is unspecified, so sort when order matters.
	fmt.Println(slices.Sorted(ages.Keys()))

	total := ages.Fold(0, func(sum int, _ string, age int) int { return sum + age })
	fmt.Println(total)

	// Output:
	// map[alan:41]
	// [ada alan]
	// 77
}

func ExampleSet() {
	team := collection.SetOf("ada", "alan", "grace")
	reviewers := collection.SetOf("grace", "linus")

	fmt.Println(team)
	fmt.Println(team.Len(), team.Contains("ada"), team.Contains("linus"))

	fmt.Println(team.Union(reviewers))
	fmt.Println(team.Intersection(reviewers))
	fmt.Println(team.Difference(reviewers))

	// Insert and Delete return new sets; the original is untouched.
	fmt.Println(team.Insert("linus").Delete("alan"), team)

	// Numbers print sorted as text; sort them yourself when the order matters.
	numbers := collection.SetOf(2, 10, 1)
	fmt.Println(numbers, slices.Sorted(numbers.All()))

	// The zero value is the empty set.
	var zero collection.Set[string]
	fmt.Println(zero.IsEmpty(), zero.Contains("ada"))

	// Output:
	// Set(ada, alan, grace)
	// 3 true false
	// Set(ada, alan, grace, linus)
	// Set(grace)
	// Set(ada, alan)
	// Set(ada, grace, linus) Set(ada, alan, grace)
	// Set(1, 10, 2) [1 2 10]
	// true false
}

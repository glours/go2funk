package collection

import (
	"fmt"
	"gotest.tools/v3/assert"
	"strconv"
	"testing"
)

var (
	_ List[int] = empty[int]{}
	_ List[int] = cons[int]{
		consHead: 10,
		consTail: Empty[int](),
		length:   1,
	}
	emptyList            = Empty[int]()
	singleElementList    = Of[int](10)
	multipleElementsList = OfSlice([]int{1, 2, 3, 4, 5})

	evenPredicate = func(value int) bool {
		return value%2 == 0
	}
)

func TestHead(t *testing.T) {
	testCases := []struct {
		name     string
		value    List[int]
		expected int
	}{
		{
			name:     "Empty List",
			value:    emptyList,
			expected: 0,
		},
		{
			name:     "Single Element List",
			value:    singleElementList,
			expected: 10,
		},
		{
			name:     "Multiple Elements List",
			value:    multipleElementsList,
			expected: 1,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.value.head(), testCase.expected, fmt.Sprintf("expected %d but value is %d", testCase.expected, testCase.value.head()))
		})
	}
}

func TestTail(t *testing.T) {
	testCases := []struct {
		name     string
		value    List[int]
		expected List[int]
	}{
		{
			name:     "Empty List",
			value:    emptyList,
			expected: emptyList,
		},
		{
			name:     "Single Element List",
			value:    singleElementList,
			expected: emptyList,
		},
		{
			name:     "Multiple Elements List",
			value:    multipleElementsList,
			expected: OfSlice[int]([]int{2, 3, 4, 5}),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.value.tail(), testCase.expected, fmt.Sprintf("expected %d but value is %d", testCase.expected, testCase.value.tail()))
		})
	}
}

func TestIsEmpty(t *testing.T) {
	testCases := []struct {
		name    string
		value   List[int]
		isEmpty bool
	}{
		{
			name:    "Empty List",
			value:   emptyList,
			isEmpty: true,
		},
		{
			name:    "Single Element List",
			value:   singleElementList,
			isEmpty: false,
		},
		{
			name:    "Multiple Elements List",
			value:   multipleElementsList,
			isEmpty: false,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.value.IsEmpty(), testCase.isEmpty, fmt.Sprintf("expected %t but value is %t", testCase.isEmpty, testCase.value.IsEmpty()))
		})
	}
}

func TestLength(t *testing.T) {
	testCases := []struct {
		name   string
		value  List[int]
		length int
	}{
		{
			name:   "Empty List",
			value:  emptyList,
			length: 0,
		},
		{
			name:   "Single Element List",
			value:  singleElementList,
			length: 1,
		},
		{
			name:   "Multiple Elements List",
			value:  multipleElementsList,
			length: 5,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.value.Length(), testCase.length, fmt.Sprintf("expected %d but value is %d", testCase.length, testCase.value.Length()))
		})
	}
}

func TestAppend(t *testing.T) {
	testCases := []struct {
		name   string
		value  List[int]
		length int
	}{
		{
			name:   "Empty List",
			value:  emptyList,
			length: 1,
		},
		{
			name:   "Single Element List",
			value:  singleElementList,
			length: 2,
		},
		{
			name:   "Multiple Elements List",
			value:  multipleElementsList,
			length: 6,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.value.Append(10)
			assert.Equal(t, result.Length(), testCase.length, fmt.Sprintf("expected %d but value is %d", testCase.length, result.Length()))
		})
	}
}

func TestAppendAll(t *testing.T) {
	testCases := []struct {
		name   string
		value  List[int]
		length int
	}{
		{
			name:   "Empty List",
			value:  emptyList,
			length: 3,
		},
		{
			name:   "Single Element List",
			value:  singleElementList,
			length: 4,
		},
		{
			name:   "Multiple Elements List",
			value:  multipleElementsList,
			length: 8,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.value.AppendAll([]int{10, 20, 30})
			assert.Equal(t, result.Length(), testCase.length, fmt.Sprintf("expected %d but value is %d", testCase.length, result.Length()))
		})
	}
}

func TestMapList(t *testing.T) {
	var mapper = func(value int) string {
		return strconv.Itoa(value)
	}
	testCases := []struct {
		name     string
		value    List[int]
		expected List[string]
	}{
		{
			name:     "Empty List",
			value:    emptyList,
			expected: Empty[string](),
		},
		{
			name:     "Single Element List",
			value:    singleElementList,
			expected: Of[string]("10"),
		},
		{
			name:     "Multiple Elements List",
			value:    multipleElementsList,
			expected: OfSlice[string]([]string{"1", "2", "3", "4", "5"}),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := MapList[int, string](testCase.value, mapper)
			assert.Equal(t, result, testCase.expected, fmt.Sprintf("expected %+v but value is %+v", testCase.expected, result))
		})
	}
}

func TestFilterList(t *testing.T) {
	testCases := []struct {
		name     string
		value    List[int]
		expected List[int]
	}{
		{
			name:     "Empty List",
			value:    emptyList,
			expected: Empty[int](),
		},
		{
			name:     "Single Element List",
			value:    singleElementList,
			expected: Of[int](10),
		},
		{
			name:     "Multiple Elements List",
			value:    multipleElementsList,
			expected: OfSlice[int]([]int{2, 4}),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.value.Filter(evenPredicate)
			assert.Equal(t, result, testCase.expected, fmt.Sprintf("expected %+v but value is %+v", testCase.expected, result))
		})
	}
}

func TestRemoveFromList(t *testing.T) {
	testCases := []struct {
		name          string
		original      List[int]
		valueToRemove int
		expected      List[int]
	}{
		{
			name:          "Empty List",
			original:      emptyList,
			valueToRemove: 10,
			expected:      Empty[int](),
		},
		{
			name:          "Single Element List with element removed",
			original:      singleElementList,
			valueToRemove: 10,
			expected:      Empty[int](),
		},
		{
			name:          "Single Element List without element removed",
			original:      singleElementList,
			valueToRemove: 5,
			expected:      Of[int](10),
		},
		{
			name:          "Multiple Elements List",
			original:      multipleElementsList,
			valueToRemove: 3,
			expected:      OfSlice[int]([]int{1, 2, 4, 5}),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.original.Remove(testCase.valueToRemove)
			assert.Equal(t, result, testCase.expected, fmt.Sprintf("expected %+v but value is %+v", testCase.expected, result))
		})
	}
}

func TestRemoveFromListWithPredicate(t *testing.T) {
	testCases := []struct {
		name      string
		original  List[int]
		predicate func(value int) bool
		expected  List[int]
	}{
		{
			name:      "Empty List",
			original:  emptyList,
			predicate: evenPredicate,
			expected:  Empty[int](),
		},
		{
			name:      "Single Element List with element removed",
			original:  singleElementList,
			predicate: evenPredicate,
			expected:  Empty[int](),
		},
		{
			name:      "Single Element List without element removed",
			original:  singleElementList,
			predicate: func(value int) bool { return value == 5 },
			expected:  Of[int](10),
		},
		{
			name:      "Multiple Elements List",
			original:  multipleElementsList,
			predicate: evenPredicate,
			expected:  OfSlice[int]([]int{1, 3, 5}),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.original.RemovePredicate(testCase.predicate)
			assert.Equal(t, result, testCase.expected, fmt.Sprintf("expected %+v but value is %+v", testCase.expected, result))
		})
	}
}

func TestInsertInList(t *testing.T) {
	testCases := []struct {
		name         string
		original     List[int]
		index        int
		expected     List[int]
		checkError   bool
		errorMessage string
	}{
		{
			name:       "Empty List index 0",
			original:   emptyList,
			index:      0,
			expected:   Of[int](7),
			checkError: false,
		},
		{
			name:         "Empty List negative index",
			original:     emptyList,
			index:        -1,
			expected:     Empty[int](),
			checkError:   true,
			errorMessage: "index out of range -1 on empty List",
		},
		{
			name:         "Empty List index > 0",
			original:     emptyList,
			index:        1,
			expected:     Empty[int](),
			checkError:   true,
			errorMessage: "index out of range 1 on empty List",
		},
		{
			name:       "Single Element List index 0",
			original:   singleElementList,
			index:      0,
			expected:   OfSlice[int]([]int{7, 10}),
			checkError: false,
		},
		{
			name:       "Single Element List index 1",
			original:   singleElementList,
			index:      1,
			expected:   OfSlice[int]([]int{10, 7}),
			checkError: false,
		},
		{
			name:         "Single Element List index > 1",
			original:     singleElementList,
			index:        2,
			expected:     Empty[int](),
			checkError:   true,
			errorMessage: "index out of range 1 on empty List",
		},
		{
			name:         "Single Element List index < 0",
			original:     singleElementList,
			index:        -1,
			expected:     Empty[int](),
			checkError:   true,
			errorMessage: "index out of range -1 on List",
		},
		{
			name:       "Multiple Elements List index 0",
			original:   multipleElementsList,
			index:      0,
			expected:   OfSlice[int]([]int{7, 1, 2, 3, 4, 5}),
			checkError: false,
		},
		{
			name:       "Multiple Elements List index 3",
			original:   multipleElementsList,
			index:      3,
			expected:   OfSlice[int]([]int{1, 2, 3, 7, 4, 5}),
			checkError: false,
		},
		{
			name:         "Multiple Elements List index out of bounds",
			original:     multipleElementsList,
			index:        7,
			expected:     Empty[int](),
			checkError:   true,
			errorMessage: "index out of range 2 on empty List",
		},
		{
			name:         "Multiple Elements List negative index",
			original:     multipleElementsList,
			index:        -1,
			expected:     Empty[int](),
			checkError:   true,
			errorMessage: "index out of range -1 on List",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := testCase.original.Insert(testCase.index, 7)
			if testCase.checkError {
				assert.Error(t, err, testCase.errorMessage, "index of range error was expected")
				if err == nil {
					t.Errorf("index of range error was expected")
				}
			} else {
				if result != testCase.expected {
					t.Errorf("expected %+v but value is %+v", testCase.expected, result)
				}
			}
		})
	}
}

func TestReverseList(t *testing.T) {
	testCases := []struct {
		name     string
		original List[int]
		expected List[int]
	}{
		{
			name:     "Empty List",
			original: emptyList,
			expected: Empty[int](),
		},
		{
			name:     "Single Element List",
			original: singleElementList,
			expected: Of[int](10),
		},
		{
			name:     "Multiple Elements List",
			original: multipleElementsList,
			expected: OfSlice[int]([]int{5, 4, 3, 2, 1}),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.original.Reverse()
			assert.Equal(t, result, testCase.expected, fmt.Sprintf("expected %+v but value is %+v", testCase.expected, result))
		})
	}
}

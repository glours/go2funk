package control

import (
	"errors"
	"strconv"
	"testing"

	"gotest.tools/v3/assert"
)

var (
	_    Option[int] = None[int]{}
	_    Option[int] = Some[int]{value: 10}
	_                = Of[int](10)
	_                = Empty[int]()
	some             = Of(10)
	none             = Empty[int]()
)

func TestIsPresent(t *testing.T) {
	assert.Assert(t, !some.IsEmpty(), "A value should be present!")
	assert.Assert(t, none.IsEmpty(), "No value should be present!")
}

func TestGetOrElse(t *testing.T) {
	assert.Equal(t, some.OrElse(20), 10, "Value should be 10 not 20")
	assert.Equal(t, none.OrElse(20), 20, "Value should be 20!")
}

func TestGetOrError(t *testing.T) {
	_, err := some.OrElseError(errors.New("should not be thrown"))
	assert.NilError(t, err)

	_, err = none.OrElseError(errors.New("should return an error"))
	assert.Error(t, err, "should return an error")
}

func TestFilter(t *testing.T) {
	var evenPredicate = func(value int) bool {
		return value%2 == 0
	}

	assert.Assert(t, !some.Filter(evenPredicate).IsEmpty(), "result of filtering should not be empty")
	assert.Assert(t, none.Filter(evenPredicate).IsEmpty(), "result of filtering should be empty")
}

func TestOptionMap(t *testing.T) {
	var mapper = func(value int) string {
		return strconv.Itoa(value)
	}

	assert.Assert(t, !MapOption[int, string](some, mapper).IsEmpty(), "result of MapOption function should not be empty")
	assert.Assert(t, MapOption[int, string](none, mapper).IsEmpty(), "result of MapOption function should be empty")
}

func TestOptionFlatMap(t *testing.T) {
	var mapper = func(value int) Option[string] {
		return Of[string](strconv.Itoa(value))
	}
	assert.Assert(t, !FlatMapOption[int, string](some, mapper).IsEmpty(), "result of FlatMapOption function should not be empty")
	assert.Assert(t, FlatMapOption[int, string](none, mapper).IsEmpty(), "result of FlatMapOption function should be empty")
}

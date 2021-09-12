package control

import (
	"errors"
	"fmt"
	"gotest.tools/v3/assert"
	"strconv"
	"testing"
)

var (
	defaultEitherError                    = errors.New("default Either error")
	_                  Either[error, int] = Right[error, int]{10}
	_                  Either[error, int] = Left[error, int]{defaultEitherError}
	right              Either[error, int] = RightOf[error, int](10)
	left               Either[error, int] = LeftOf[error, int](defaultEitherError)
)

func TestIsRight(t *testing.T) {
	assert.Assert(t, right.IsRight(), "should be a Right not Left")
	assert.Assert(t, !left.IsRight(), "should be a Left not Right")
}

func TestIsLeft(t *testing.T) {
	assert.Assert(t, !right.IsLeft(), "should be a Right not Left")
	assert.Assert(t,  left.IsLeft(), "should be a Left not Right")
}

func TestSwap(t *testing.T) {
	assert.Assert(t, right.Swap().IsLeft(), "should be a Left not Right")
	assert.Assert(t, left.Swap().IsRight(), "should be a Right not Left")
}

func TestEitherOrElse(t *testing.T) {
	assert.Assert(t, right.OrElse(left).IsRight(), "should be a Right not Left")
	assert.Assert(t, left.OrElse(right).IsRight(), "should be a Right not Left")
}

func TestEitherGetOrElse(t *testing.T) {
	assert.Equal(t, right.GetOrElse(20), 10, "value should be 10")
	assert.Equal(t, left.GetOrElse(20), 20, "value should be 20")
}

func TestEitherFilterOrElse(t *testing.T) {
	transform := func(value int) error {
		return fmt.Errorf("doesn't pass the EvenPredicate")
	}
	assert.Assert(t, right.FilterOrElse(EvenPredicate, transform).IsRight(), "should not be a Left")
	assert.Assert(t, left.FilterOrElse(EvenPredicate, transform).IsLeft(), "should not be a Right")

	odd := RightOf[error, int](11)
	assert.Assert(t, odd.FilterOrElse(EvenPredicate, transform).IsLeft(), "should not be a Right")
}

func TestEitherFilter(t *testing.T) {
	assert.Assert(t, !right.Filter(EvenPredicate).IsEmpty(), "should be a Some of Either")
	assert.Assert(t, left.Filter(EvenPredicate).IsEmpty(), "should be a Empty of Either")

	odd := RightOf[error, int](11)
	assert.Assert(t, odd.Filter(EvenPredicate).IsEmpty(), "should be a Empty of Either")
}

func TestMapEither(t *testing.T) {
	var mapper = func(value int) string {
		return strconv.Itoa(value)
	}
	var mapRight Either[error, string] = MapEither(right, mapper)
	assert.Equal(t, mapRight.GetOrElse("good"), "10", "value should be 10")

	var mapLeft Either[error, string] = MapEither[error, int, string](left, mapper)
	assert.Assert(t, mapLeft.IsLeft(), "should be an Left Either")
}

func TestFlatMapEither(t *testing.T) {
	var mapper = func(value int) Either[error, string] {
		return RightOf[error, string](strconv.Itoa(value))
	}
	var mapRight Either[error, string] = FlatMapEither(right, mapper)
	assert.Equal(t, mapRight.GetOrElse("good"), "10", "value should be 10")

	var mapLeft Either[error, string] = FlatMapEither[error, int, string](left, mapper)
	assert.Assert(t, mapLeft.IsLeft(), "should be an Left Either")
}

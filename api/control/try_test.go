package control

import (
	"errors"
	"gotest.tools/v3/assert"
	"strconv"
	"testing"
)

var (
	defaultTryError            = errors.New("default Try error")
	nilCauseError              = errors.New("error cause should not be nil")
	nilPredicateError          = errors.New("predicate should not be nil")
	_                 Try[int] = Success[int]{10}
	_                 Try[int] = Failure[int]{cause: defaultTryError}
	success                    = SuccessOf[int](10)
	failure                    = FailureOf[int](defaultTryError)
	EvenPredicate              = func(value int) bool {
		return value%2 == 0
	}
)

func TestIsFailure(t *testing.T) {
	assert.Assert(t, !success.IsFailure(), "success should not be a failure")
	assert.Assert(t, failure.IsFailure(), "failure should not be a success")
}

func TestTryOf(t *testing.T) {
	assert.Assert(t, !TryOf(func() (int, error) { return 10, nil }).IsFailure(), "should not be a failure")
	assert.Assert(t, TryOf(func() (int, error) { return 0, defaultTryError }).IsFailure(), "should not be a success")
}

func TestOrElse(t *testing.T) {
	assert.Equal(t, success.OrElse(20), 10, "value should be 10")
	assert.Equal(t, failure.OrElse(20), 20, "value should be 20")
}

func TestOrElseCause(t *testing.T) {
	value, err := success.OrElseCause()
	assert.NilError(t, err, err)
	assert.Equal(t, value, 10, err)

	_, err = failure.OrElseCause()
	assert.Error(t, err, "default Try error", "should return default cause")
}

func TestTryFilterCheckParam(t *testing.T) {
	_, err := success.Filter(nil, defaultTryError).OrElseCause()
	assert.Error(t, err, nilPredicateError.Error(), "should return a Try with nil predicate cause")

	_, err = success.Filter(EvenPredicate, nil).OrElseCause()
	assert.Error(t, err, nilCauseError.Error(), "should return a Try with nil cause error")

	_, err = failure.Filter(nil, defaultTryError).OrElseCause()
	assert.Error(t, err, nilPredicateError.Error(), "should return a Try with nil predicate cause")

	_, err = failure.Filter(EvenPredicate, nil).OrElseCause()
	assert.Error(t, err, nilCauseError.Error(), "should return a Try with nil cause error")
}

func TestTryFilter(t *testing.T) {
	assert.Assert(t, !success.Filter(EvenPredicate, defaultTryError).IsFailure(), "should not be a failure")
	assert.Assert(t, failure.Filter(EvenPredicate, defaultTryError).IsFailure(), "should not be a success")
}

func TestTryMap(t *testing.T) {
	var mapper = func(value int) string {
		return strconv.Itoa(value)
	}
	assert.Assert(t, MapTry[int, string](failure, mapper).IsFailure(), "result of MapTry function should be a failure")
	assert.Assert(t, !MapTry[int, string](success, mapper).IsFailure(), "result of MapTry function should be a success")
}

func TestTryFlatMap(t *testing.T) {
	var mapper = func(value int) Try[string] {
		return SuccessOf[string](strconv.Itoa(value))
	}
	assert.Assert(t, FlatMapTry[int, string](failure, mapper).IsFailure(), "result of MapTry function should be a failure")
	assert.Assert(t, !FlatMapTry[int, string](success, mapper).IsFailure(), "result of MapTry function should be a success")
}

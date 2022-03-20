package api

import (
	"fmt"
	"gotest.tools/v3/assert"
	"strconv"
	"testing"
)

var (
	_          = Pair[int, string]{10, "ten"}
	pair       = NewPair[int, string](10, "ten")
	mapperLeft = func(value int) string {
		return strconv.Itoa(value)
	}
	mapperRight = func(value string) []byte {
		return []byte(value)
	}
)

func TestGetLeft(t *testing.T) {
	assert.Equal(t, pair.GetLeft(), 10, fmt.Sprintf("value should be 10 but is %d", pair.GetLeft()))
}

func TestGetRight(t *testing.T) {
	assert.Equal(t, pair.GetRight(), "ten", fmt.Sprintf("value should be 'ten' but is '%s'", pair.GetRight()))
}

func TestMapLeftPair(t *testing.T) {
	mappedLeft := MapLeftPair(pair, mapperLeft)
	assert.Equal(t, mappedLeft.GetLeft(), "10", fmt.Sprintf("value should be '10' but is '%s'", mappedLeft.GetLeft()))
}

func TestMapRightPair(t *testing.T) {
	mappedRight := MapRightPair(pair, mapperRight)
	assert.Equal(t, string(mappedRight.GetRight()), "ten", fmt.Sprintf("value should be a []array representing 'ten' string but is %s", mappedRight.GetRight()))
}

func TestMapPair(t *testing.T) {
	mapped := MapPair(pair, mapperLeft, mapperRight)
	assert.Equal(t, mapped.GetLeft(), "10", fmt.Sprintf("value should be '10' but is '%s'", mapped.GetLeft()))

	assert.Equal(t, string(mapped.GetRight()), "ten", fmt.Sprintf("value should be a []array representing 'ten' string but is %s", mapped.GetRight()))
}

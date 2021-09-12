package api

import (
	"strconv"
	"testing"
)

var (
	_          Pair[int, string] = Pair[int, string]{10, "ten"}
	pair                         = NewPair[int, string](10, "ten")
	mapperLeft                   = func(value int) string {
		return strconv.Itoa(value)
	}
	mapperRight = func(value string) []byte {
		return []byte(value)
	}
)

func TestGetLeft(t *testing.T) {
	if pair.GetLeft() != 10 {
		t.Errorf("value should be 10 but is %d", pair.GetLeft())
	}
}

func TestGetRight(t *testing.T) {
	if pair.GetRight() != "ten" {
		t.Errorf("value should be 'ten' but is '%s'", pair.GetRight())
	}
}

func TestMapLeftPair(t *testing.T) {
	mappedLeft := MapLeftPair(pair, mapperLeft)
	if mappedLeft.GetLeft() != "10" {
		t.Errorf("value should be '10' but is '%s'", mappedLeft.GetLeft())
	}
}

func TestMapRightPair(t *testing.T) {
	mappedRight := MapRightPair(pair, mapperRight)
	if string(mappedRight.GetRight()) != "ten" {
		t.Errorf("value should be a []array representing 'ten' string but is %s", mappedRight.GetRight())
	}
}

func TestMapPair(t *testing.T) {
	mapped := MapPair(pair, mapperLeft, mapperRight)
	if mapped.GetLeft() != "10" {
		t.Errorf("value should be '10' but is '%s'", mapped.GetLeft())
	}

	if string(mapped.GetRight()) != "ten" {
		t.Errorf("value should be a []array representing 'ten' string but is %s", mapped.GetRight())
	}
}

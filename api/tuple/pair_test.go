package tuple_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/glours/go2funk/api/tuple"
)

func TestNewAndAccessors(t *testing.T) {
	p := tuple.New("ten", 10)

	if p.Left() != "ten" || p.Right() != 10 {
		t.Errorf("pair = (%q, %d), want (\"ten\", 10)", p.Left(), p.Right())
	}

	left, right := p.Unpack()
	if left != "ten" || right != 10 {
		t.Errorf("Unpack = (%q, %d), want (\"ten\", 10)", left, right)
	}
}

func TestZeroValue(t *testing.T) {
	var p tuple.Pair[string, int]

	if p.Left() != "" || p.Right() != 0 {
		t.Error("the zero value of Pair must hold the zero values of both sides")
	}
}

func TestSwap(t *testing.T) {
	swapped := tuple.New("ten", 10).Swap()

	if swapped.Left() != 10 || swapped.Right() != "ten" {
		t.Errorf("Swap = (%d, %q), want (10, \"ten\")", swapped.Left(), swapped.Right())
	}
}

func TestMapSides(t *testing.T) {
	p := tuple.New("ten", 10)

	var right tuple.Pair[string, string] = p.MapRight(strconv.Itoa)
	if right.Right() != "10" {
		t.Errorf("MapRight = %q, want %q", right.Right(), "10")
	}

	var left tuple.Pair[int, int] = p.MapLeft(func(s string) int { return len(s) })
	if left.Left() != 3 {
		t.Errorf("MapLeft = %d, want 3", left.Left())
	}
}

func TestMapBoth(t *testing.T) {
	var both tuple.Pair[string, bool] = tuple.New("ten", 10).
		Map(strings.ToUpper, func(value int) bool { return value > 5 })

	if both.Left() != "TEN" || !both.Right() {
		t.Errorf("Map = (%q, %v), want (\"TEN\", true)", both.Left(), both.Right())
	}
}

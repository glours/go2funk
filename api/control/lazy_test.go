package control_test

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/glours/go2funk/api/control"
)

func TestLazyIsNotEvaluatedUntilRead(t *testing.T) {
	calls := 0
	lazy := control.NewLazy(func() int { calls++; return 10 })

	if lazy.IsEvaluated() {
		t.Error("a Lazy must not be evaluated before Get")
	}
	if calls != 0 {
		t.Errorf("the function ran %d times before Get, want 0", calls)
	}

	if got := lazy.Get(); got != 10 {
		t.Errorf("Get = %d, want 10", got)
	}
	if !lazy.IsEvaluated() {
		t.Error("a Lazy must report itself evaluated after Get")
	}
	if calls != 1 {
		t.Errorf("the function ran %d times, want 1", calls)
	}
}

func TestLazyMemoizes(t *testing.T) {
	calls := 0
	lazy := control.NewLazy(func() int { calls++; return calls })

	first, second, third := lazy.Get(), lazy.Get(), lazy.Get()
	if first != 1 || second != 1 || third != 1 {
		t.Errorf("Get returned %d, %d, %d; want 1, 1, 1", first, second, third)
	}
	if calls != 1 {
		t.Errorf("the function ran %d times, want 1", calls)
	}
}

// The zero value behaves like Rust's LazyCell::default(): no computation is
// attached, so it yields the zero value of T. It must never panic.
func TestLazyZeroValue(t *testing.T) {
	var number control.Lazy[int]
	if got := number.Get(); got != 0 {
		t.Errorf("zero Lazy[int].Get = %d, want 0", got)
	}
	if !number.IsEvaluated() {
		t.Error("a zero Lazy has nothing to evaluate, so it is already evaluated")
	}

	type custom struct {
		Name  string
		Count int
	}
	var value control.Lazy[custom]
	if got := value.Get(); got != (custom{}) {
		t.Errorf("zero Lazy[custom].Get = %+v, want the zero struct", got)
	}
}

func TestLazyWithNilFunction(t *testing.T) {
	lazy := control.NewLazy[int](nil)

	if got := lazy.Get(); got != 0 {
		t.Errorf("NewLazy(nil).Get = %d, want 0", got)
	}
	if !lazy.IsEvaluated() {
		t.Error("NewLazy(nil) has nothing to evaluate")
	}
}

func TestDelayIsNewLazy(t *testing.T) {
	if got := control.Delay(func() string { return "ada" }).Get(); got != "ada" {
		t.Errorf("Delay(...).Get = %q, want %q", got, "ada")
	}
}

func TestLazyMapStaysLazyAndChangesType(t *testing.T) {
	sourceCalls, mapperCalls := 0, 0

	source := control.NewLazy(func() int { sourceCalls++; return 10 })
	var mapped control.Lazy[string] = source.Map(func(value int) string {
		mapperCalls++
		return strconv.Itoa(value)
	})

	if sourceCalls != 0 || mapperCalls != 0 {
		t.Errorf("Map evaluated eagerly: source=%d mapper=%d", sourceCalls, mapperCalls)
	}
	if mapped.IsEvaluated() {
		t.Error("the mapped Lazy must not be evaluated before Get")
	}

	if got := mapped.Get(); got != "10" {
		t.Errorf("Get = %q, want %q", got, "10")
	}
	if got := mapped.Get(); got != "10" {
		t.Errorf("second Get = %q, want %q", got, "10")
	}
	if sourceCalls != 1 || mapperCalls != 1 {
		t.Errorf("ran source=%d mapper=%d times, want 1 and 1", sourceCalls, mapperCalls)
	}
}

func TestLazyFlatMap(t *testing.T) {
	calls := 0
	source := control.NewLazy(func() int { return 10 })

	flat := source.FlatMap(func(value int) control.Lazy[string] {
		return control.NewLazy(func() string { calls++; return strconv.Itoa(value * 2) })
	})

	if flat.IsEvaluated() || calls != 0 {
		t.Error("FlatMap must not evaluate anything before Get")
	}
	if got := flat.Get(); got != "20" {
		t.Errorf("FlatMap.Get = %q, want %q", got, "20")
	}
	if calls != 1 {
		t.Errorf("the inner function ran %d times, want 1", calls)
	}
}

// Copying a Lazy shares the memoized result rather than restarting it.
func TestLazyCopyShareTheResult(t *testing.T) {
	calls := 0
	original := control.NewLazy(func() int { calls++; return 10 })

	copied := original
	if got := copied.Get(); got != 10 {
		t.Errorf("copy.Get = %d, want 10", got)
	}
	if got := original.Get(); got != 10 {
		t.Errorf("original.Get = %d, want 10", got)
	}
	if calls != 1 {
		t.Errorf("the function ran %d times across the copies, want 1", calls)
	}
}

// Run with -race: concurrent readers must see one evaluation and one value.
func TestLazyIsSafeUnderConcurrency(t *testing.T) {
	const readers = 64

	var calls atomic.Int32
	lazy := control.NewLazy(func() int {
		calls.Add(1)
		return 42
	})

	var start, done sync.WaitGroup
	start.Add(1)
	done.Add(readers)
	results := make([]int, readers)

	for i := range readers {
		go func() {
			defer done.Done()
			start.Wait()
			results[i] = lazy.Get()
		}()
	}
	start.Done()
	done.Wait()

	if got := calls.Load(); got != 1 {
		t.Errorf("the function ran %d times, want exactly 1", got)
	}
	for i, value := range results {
		if value != 42 {
			t.Fatalf("reader %d saw %d, want 42", i, value)
		}
	}
}

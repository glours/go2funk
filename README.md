# Go2Funk

Go2Funk is a pet project exploring what Go generics make possible: implementing
purely functional structures — `Option`, `Try`, `Either`, `List`, `Pair` — the way
[Vavr](https://vavr.io) does for Java.

**No runtime dependencies.** The `api/` packages import nothing outside the Go
standard library, and that is a hard constraint rather than a preference.

> **Status** — the project is being picked back up. Go 1.27 (August 2026) added
> generic methods, which lifts the language limitation the current API was
> designed around, so a rework is under way. Expect the API below to change.

## Requirements

Go 1.20 or later.

## Install

```sh
go get github.com/glours/go2funk
```

## Usage

Every snippet below is backed by a runnable `Example` test, so it compiles and its
output is verified by `go test ./...`.

### Option

An `Option[T]` is either `Some` (a value is present) or `None` (it is not).

```go
import "github.com/glours/go2funk/api/control"

empty := control.Empty[int]()
some := control.Of(10)

fmt.Println(empty.OrElse(5))  // 5
fmt.Println(some.OrElse(5))   // 10

isEven := func(value int) bool { return value%2 == 0 }
fmt.Println(some.Filter(isEven).IsEmpty())  // false

asString := control.MapOption(some, strconv.Itoa)
fmt.Println(asString.OrElse("none"))  // 10
```

`OrElseError` turns an empty `Option` into a Go error:

```go
missing := errors.New("no value")

_, err := control.Empty[int]().OrElseError(missing)
fmt.Println(err)  // no value

value, err := control.Of(10).OrElseError(missing)
fmt.Println(value, err)  // 10 <nil>
```

`Map` and `FlatMap` are package-level functions rather than methods
(`MapOption`, `FlatMapOption`) because a method cannot introduce a new type
parameter before Go 1.27.

See [`api/control/example_test.go`](./api/control/example_test.go) and
[`api/control/option_test.go`](./api/control/option_test.go).

### Try

A `Try[A]` is either a `Success` carrying a value or a `Failure` carrying an error.

```go
import "github.com/glours/go2funk/api/control"

boom := errors.New("boom")

success := control.SuccessOf(10)
failure := control.FailureOf[int](boom)

fmt.Println(success.IsFailure(), failure.IsFailure())  // false true
fmt.Println(success.OrElse(5), failure.OrElse(5))      // 10 5

_, err := failure.OrElseCause()
fmt.Println(err)  // boom

fmt.Println(control.TryOf(func() (int, error) { return 10, nil }).IsFailure())   // false
fmt.Println(control.TryOf(func() (int, error) { return 0, boom }).IsFailure())   // true
```

See [`api/control/example_test.go`](./api/control/example_test.go) and
[`api/control/try_test.go`](./api/control/try_test.go).

### Either

An `Either[L, R]` holds one of two types. By convention `Right` carries the
expected value and `Left` the alternative one.

```go
import "github.com/glours/go2funk/api/control"

boom := errors.New("boom")
noError := errors.New("no error")

right := control.RightOf[error](10)
left := control.LeftOf[error, int](boom)

fmt.Println(right.IsRight(), left.IsLeft())                      // true true
fmt.Println(right.GetOrElse(20), left.GetOrElse(20))             // 10 20
fmt.Println(right.GetLeftOrElse(noError), left.GetLeftOrElse(noError))  // no error boom

asString := control.MapEither(right, strconv.Itoa)
fmt.Println(asString.GetOrElse("none"))  // 10

fmt.Println(right.Swap().GetLeftOrElse(0))  // 10
```

`MapEither` and `FlatMapEither` operate on the `Right` side; a `Left` passes
through unchanged.

See [`api/control/example_test.go`](./api/control/example_test.go) and
[`api/control/either_test.go`](./api/control/either_test.go).

### List

An immutable, persistent linked list.

```go
import "github.com/glours/go2funk/api/collection"

list := collection.OfSlice([]int{1, 2, 3, 4, 5})
list = list.Append(6)
fmt.Println(list.Length())  // 6

isEven := func(value int) bool { return value%2 == 0 }
evens := list.Filter(isEven)
fmt.Println(evens.Length())  // 3

asStrings := collection.MapList(evens, strconv.Itoa)
fmt.Println(asStrings.Length(), asStrings.IsEmpty())  // 3 false
```

`Insert` returns an error when the index is out of range:

```go
list := collection.OfSlice([]int{1, 2, 4})

inserted, err := list.Insert(2, 3)
fmt.Println(inserted.Length(), err)  // 4 <nil>

_, err = list.Insert(42, 3)
fmt.Println(err != nil)  // true
```

> **Known limitation** — `List` currently exposes no exported way to read its
> elements back out: there is no `Head`, `Get`, `ToSlice` or iterator. This is
> part of the planned rework.

See [`api/collection/example_test.go`](./api/collection/example_test.go) and
[`api/collection/list_test.go`](./api/collection/list_test.go).

### Pair

A two-element product type, with independent mappers for each side.

```go
import "github.com/glours/go2funk/api"

pair := api.NewPair("ten", 10)
fmt.Println(pair.GetLeft(), pair.GetRight())  // ten 10

asString := api.MapRightPair(pair, strconv.Itoa)
fmt.Println(asString.GetRight())  // 10

both := api.MapPair(pair, strings.ToUpper, func(value int) bool { return value > 5 })
fmt.Println(both.GetLeft(), both.GetRight())  // TEN true
```

See [`api/example_test.go`](./api/example_test.go) and
[`api/pair_test.go`](./api/pair_test.go).

## Development

```sh
go build ./...          # build
go test ./...           # run the test suite, examples included
go test -race -cover ./...
go vet ./...            # static checks
gofmt -l .              # must print nothing
golangci-lint run ./... # full lint, config in .golangci.yml
```

The no-runtime-dependency rule is enforced two ways: by `depguard` in
`.golangci.yml`, which rejects any non-standard-library import under `api/`, and
by this command, which must print nothing:

```sh
go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./... | grep -v go2funk
```

Both run in CI, along with the build and test matrix.

### Contributing

[`AGENTS.md`](AGENTS.md) holds the development rules for this repository — no
runtime dependencies, test-driven development, black-box tests, and the API
design decisions. They apply to humans and coding agents alike. `CLAUDE.md` is a
symlink to it.

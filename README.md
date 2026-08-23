# Go2Funk

Go2Funk is a pet project exploring what Go generics make possible: implementing
purely functional structures — `Option`, `Either`, `Result`, `List`, `Pair` — the
way [Vavr](https://vavr.io) does for Java.

**No runtime dependencies.** The `api/` packages import nothing outside the Go
standard library, and that is a hard constraint rather than a preference.

Go 1.27 added generic methods, so `Map` can change the type it carries *and* stay
chainable:

```go
control.Some(user).Map(User.Name).Filter(nonEmpty).Map(strings.ToUpper).OrElse("anonymous")
```

No other Go library does both today: `Option` is a plain struct, there is no
interface and no boxing, and the zero value is `None` rather than a nil panic.

> **Status** — `api/collection` still has the older interface-based design and is
> next in line. Expect it to change.

## Requirements

Go 1.27 or later. Generic methods are the whole point of the API, and they do not
exist before that.

## Install

```sh
go get github.com/glours/go2funk
```

## Usage

Every snippet below is backed by a runnable `Example` test, so it compiles and its
output is verified by `go test ./...`.

### Option

`Option[T]` is either `Some`, holding a value, or `None`. Its zero value is
`None`, so it needs no initialisation.

```go
import "github.com/glours/go2funk/api/control"

some := control.Some(10)
none := control.None[int]()

fmt.Println(some.OrElse(5), none.OrElse(5))  // 10 5

// Map is a method and changes the type it carries.
fmt.Println(some.Map(strconv.Itoa).Map(strings.ToUpper).OrElse("none"))  // 10

isEven := func(value int) bool { return value%2 == 0 }
fmt.Println(some.Filter(isEven).IsDefined())  // true

var zero control.Option[int]
fmt.Println(zero.IsEmpty())  // true — no panic
```

Interop with the Go idioms it replaces:

```go
value, ok := counts["ten"]
found := control.FromTuple(value, ok)      // "comma ok" -> Option

value, err := found.OrElseError(errors.New("no value"))  // Option -> (T, error)

found.ToSlice()      // []int{10}
found.ToPointer()    // *int, nil when empty
control.FromPointer(p)
```

Also available: `Get`, `OrElseGet`, `Or`, `FlatMap`, `Fold`, `ForEach`.

See [`api/control/example_test.go`](./api/control/example_test.go).

### Either

`Either[L, R]` holds one of two values. By convention `Right` carries the expected
one, so `Map`, `FlatMap` and `FilterOrElse` work on the right side and let a
`Left` through untouched, value intact.

```go
right := control.Right[string](10)
left := control.Left[string, int]("nope")

fmt.Println(right.OrElse(20), left.OrElse(20))         // 10 20
fmt.Println(right.Map(strconv.Itoa).OrElse("none"))    // 10
fmt.Println(left.Map(strconv.Itoa).LeftOrElse("?"))    // nope

fmt.Println(right.Fold(
    func(s string) string { return "left: " + s },
    func(v int) string { return "right: " + strconv.Itoa(v) },
))  // right: 10
```

Also available: `Get`, `GetLeft`, `Swap`, `MapLeft`, `Or`, `OrElseGet`, `ForEach`,
`ToOption`.

### Result

`Result[T]` is a type alias for `Either[error, T]`, so it *is* an `Either` and
inherits every one of its methods.

```go
parse := func(s string) control.Result[int] {
    return control.Try(func() (int, error) { return strconv.Atoi(s) })
}

fmt.Println(parse("42").Map(func(v int) int { return v * 2 }).OrElse(-1))  // 84

// The cause survives Map; Unwrap hands it back to the Go idiom.
_, err := control.Unwrap(parse("nope").Map(strconv.Itoa))
fmt.Println(err)  // strconv.Atoi: parsing "nope": invalid syntax
```

`Ok`, `Err`, `Try` and `Unwrap` are package-level functions rather than methods
because a type alias cannot declare methods of its own.

### Pair

```go
import "github.com/glours/go2funk/api/tuple"

pair := tuple.New("ten", 10)
fmt.Println(pair.Left(), pair.Right())  // ten 10

left, right := pair.Unpack()
fmt.Println(pair.MapRight(strconv.Itoa).Right())  // 10
fmt.Println(pair.Swap().Left())                   // 10
```

See [`api/tuple/example_test.go`](./api/tuple/example_test.go).

### List

An immutable, persistent linked list. This package has not been reworked yet: it
is still interface-based, and `Map` is a package-level function.

```go
import "github.com/glours/go2funk/api/collection"

list := collection.OfSlice([]int{1, 2, 3, 4, 5})
list = list.Append(6)
fmt.Println(list.Length())  // 6

isEven := func(value int) bool { return value%2 == 0 }
asStrings := collection.MapList(list.Filter(isEven), strconv.Itoa)
fmt.Println(asStrings.Length(), asStrings.IsEmpty())  // 3 false
```

> **Known limitation** — `List` exposes no way to read its elements back: there is
> no `Head`, `Get`, `ToSlice` or iterator. This is the subject of the next rework.

See [`api/collection/example_test.go`](./api/collection/example_test.go).

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

### Releases

Pushing a `v*` tag triggers GoReleaser, which publishes the GitHub release, its
notes and the source archive. Config is in `.goreleaser.yaml`; check it with
`goreleaser check` and dry-run with `goreleaser release --snapshot --clean`.

### Contributing

[`AGENTS.md`](AGENTS.md) holds the development rules for this repository — no
runtime dependencies, test-driven development, black-box tests, and the API
design decisions. They apply to humans and coding agents alike. `CLAUDE.md` is a
symlink to it.

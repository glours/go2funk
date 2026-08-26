# Go2Funk

Go2Funk is a pet project exploring what Go generics make possible: implementing
purely functional structures — `Option`, `Either`, `Result`, `Lazy`, `List`,
`Pair` — the way [Vavr](https://vavr.io) does for Java.

**No runtime dependencies.** The `api/` packages import nothing outside the Go
standard library, and that is a hard constraint rather than a preference.

Go 1.27 added generic methods, so `Map` can change the type it carries *and* stay
chainable:

```go
control.Some(user).Map(User.Name).Filter(nonEmpty).Map(strings.ToUpper).OrElse("anonymous")
```

No other Go library does both today: `Option` is a plain struct, there is no
interface and no boxing, and the zero value is `None` rather than a nil panic.

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

`Option` also encodes to and from JSON, which the plain struct cannot do — an
`Option` field without it marshals to `{}` whether it holds a value or not:

```go
type Profile struct {
    Name     string                 `json:"name"`
    Nickname control.Option[string] `json:"nickname"`
    Age      control.Option[int]    `json:"age,omitzero"`
}

// {"name":"ada","nickname":"countess"}   — age is dropped, nickname is null when None
```

`None` encodes as `null`. Decoding an explicit `null` gives `None`; a missing key
leaves the field alone, which for a fresh value means `None`. Decoding into an
`Option` that already holds a value merges into it, as it would for a plain
field of the same type.

Use the `omitzero` tag (Go 1.24+) to leave the field out entirely — it behaves
the same with both encoders. Avoid `omitempty` here: `encoding/json` emits `null`
while `encoding/json/v2` drops the field, so the shape would differ depending on
which one the caller uses.

`Option` implements both forms of the marshaling contract: the byte-slice one
(`json.Marshaler`, which is an alias for `encoding/json/v2.Marshaler`) and the
streaming one (`MarshalJSONTo`). Both encoders prefer the streaming form, which
is what lets the caller's options reach the value inside the `Option`. Code that
dispatches on `json.Marshaler` reaches the byte-slice form instead, and that form
cannot see those options — notably it always escapes HTML, as `encoding/json`
does by default.

> This support imports `encoding/json/v2`, which Go guards behind the `jsonv2`
> experiment. Building with `GOEXPERIMENT=nojsonv2` therefore fails on the whole
> `control` package, not just its JSON methods. That flag is a transitional
> escape hatch for the v2 rollout and go2funk does not work around it.

Any value that encodes as `null` — a nil pointer, slice, map or interface, or a
nested `None` — decodes back to `None`, so the outer "a value is present" bit is
lost for those.


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

### Lazy

`Lazy[T]` defers a computation until its result is first read, then remembers it.
The computation runs at most once, however many goroutines ask for it.

```go
config := control.NewLazy(loadConfig)  // nothing has run yet

config.Get()          // runs loadConfig
config.Get()          // returns the remembered result
config.IsEvaluated()  // true
```

`Map` and `FlatMap` stay lazy — chaining them runs nothing until the result is
read:

```go
report := control.NewLazy(fetchRows).Map(summarise).Map(render)
// still nothing has run
report.Get()
```

`Delay` is an alias for `NewLazy`, under the name functional languages give it.

A `Lazy` is a value: copying one shares the memoised result rather than
restarting the computation. Its zero value has no computation attached and
yields the zero value of `T`, the way Rust's `LazyCell::default()` does.

`Lazy` carries no notion of absence or failure, and does not need to: use
`Lazy[Option[T]]` when "not computed yet" has to be told apart from "computed to
the zero value", and `Lazy[Result[T]]` when the computation can fail.

See [`api/control/example_test.go`](./api/control/example_test.go).

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

An immutable, persistent singly linked list. The zero value is the empty list.
`Prepend`, `Head`, `Tail`, `Length` and `IsEmpty` are O(1); everything that has
to walk the list is O(n) and says so in its godoc.

```go
import "github.com/glours/go2funk/api/collection"

list := collection.Of(1, 2, 3, 4, 5)

fmt.Println(list.Length(), list.Head().OrElse(-1))  // 5 1

isEven := func(value int) bool { return value%2 == 0 }
fmt.Println(list.Filter(isEven).Map(strconv.Itoa))  // List(2, 4)

// Prepend is O(1) and the original list is untouched.
fmt.Println(list.Prepend(0), list)  // List(0, 1, 2, 3, 4, 5) List(1, 2, 3, 4, 5)
```

`Head` and `Get` return an `Option`, so reading an element out of range is a
value rather than a panic or a second return.

`All` returns an `iter.Seq[T]`, which makes a `List` usable directly in a
`for range` loop and with anything else that speaks the Go iterator protocol:

```go
for value := range list.All() {
    fmt.Println(value)
}

collection.Collect(seq)  // iter.Seq[T] -> List[T]
```

`Fold` combines from the left and can change the type:

```go
sum := list.Fold(0, func(acc, value int) int { return acc + value })
joined := list.Fold("", func(acc string, value int) string { return acc + strconv.Itoa(value) })
```

Also available: `Tail`, `Append`, `AppendAll`, `Reverse`, `Insert`, `FlatMap`,
`ForEach`, `ToSlice`, `String`, and `collection.Remove` for lists of a
comparable type.

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

# Go2Funk

Go2Funk is a pet project to test Go2 Generics and see how this new feature help us to implement some functional
structures easily.  
The idea is to keep this repo as much as possible dependency free.

## How to compile

You should follow the [golang process](https://go.googlesource.com/go/+/refs/heads/dev.go2go/README.go2go.md) to get the
go2go tool and be able to compile the project

## Build

use `go tool go2go build` or if you
use [Goland then configure external tools](https://www.jetbrains.com/help/go/how-to-use-type-parameters-for-generic-programming.html)

## Run tests

use `go tool go2go test` or if you
use [Goland then configure external tools](https://www.jetbrains.com/help/go/how-to-use-type-parameters-for-generic-programming.html)

## Usage

### Collection types

#### List

```go
myIntArray := []int{1, 2, 3, 4, 5}
myList := OfSlice[int](myIntArray)

myList = myList.Append(6)
fmt.Println(myList.Length())

var mapper = func (value int) string { return strconv.Itoa(value) }

var evenPredicate = func (value int) bool { return value % 2 == 0 }

myList = MapList[int, string](myList.Filter(evenPredicate), mapper); 
```

For more usage details check [tests](./api/collection/list_test.go2).

### Control types

#### Option

```go
empytOption := Empty[int]()
someOption := Of[int](10)

fmt.Println(emptyOption.GetOrElse(5)) // Print 5
fmt.Println(someOption.GetOrElse(5)) // Print 10

var evenPredicate = func (value int) bool { return value % 2 == 0 }
fmt.Println(some.Filter(eventPredicate).IsEmpty()) #Print false

var mapper = func (value int) string { return strconv.Itoa(value) }
fmt.Println(MapOption(someOption, mapper)) // Print "10"
```

For more usage details check [tests](./api/control/option_test.go2).

#### Try

```go
defaultTryError := errors.New("default Try error")

failure := FailureOf[int](defaultTryError)
success := SuccessOf[int](10)

fmt.Println(failure.GetOrElse(5)) // Print 5
fmt.Println(someOption.GetOrElse(5)) // Print 10

fmt.Println(failure.GetOrElseCause()) // Print "default Try error"
fmt.Println(someOption.GetOrElseCause()) // Print 10

successLambda := func ()(int, error) { return 10, nil})
failureLambda := func ()(int, error) {return 0, defaultTryError }

fmt.Println(TryOf(failureLambda).IsFailure()) // Print true
fmt.Println(TryOf(successLambda).IsFailure()) // Print false
```

For more usage details check [tests](./api/control/try_test.go2).

#### Either

```go
defaultEitherError = errors.New("default Either error")
right Either[error, int] = RightOf[error, int](10)
left Either[error, int] = LeftOf[error, int](defaultEitherError)

fmt.Println(right.GetOrElse(20)) // Print 10
fmt.Println(left.GetOrElse(20)) // Print 20

fmt.Println(right.GetLeftOrElse(error.New("new error"))) // Print "new error"
fmt.Println(left.GetOrElse(error.New("new error"))) // Print "default Either error"

var mapper = func (value int) string { return strconv.Itoa(value) }
var mapRight Either[error, string] = MapEither(right, mapper)
fmt.Println(mapRight.GetOrElse("good")) // Print "10"
```

For more usage details check [tests](./api/control/either_test.go2).
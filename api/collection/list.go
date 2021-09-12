// Package collection provides purely functional collections
package collection

import (
	"fmt"
	"reflect"
)

// List is a immutable interface of List collection.
type List[T any] interface {
	head() T
	tail() List[T]
	IsEmpty() bool
	Append(value T) List[T]
	AppendAll(values []T) List[T]
	Length() int
	Filter(func(T) bool) List[T]
	Remove(value T) List[T]
	RemovePredicate(func(T) bool) List[T]
	Insert(int, T) (List[T], error)
	Reverse() List[T]
}

// MapList maps the elements of the List[T] to elements of a new type U preserving their order, if any.
func MapList[T any, U any](list List[T], mapper func(T) U) List[U] {
	result := Empty[U]()
	if list.IsEmpty() {
		return result
	}
	return newCons[U](mapper(list.head()), MapList(list.tail(), mapper))
}

// Empty provide an empty List which could contains elements of T type.
func Empty[T any]() List[T] {
	return empty[T]{}
}

// Of provide a single element List of type T.
func Of[T any](value T) List[T] {
	return newCons[T](value, Empty[T]())
}

// OfSlice provide a List which contains the elements of type T provided by the array passed as parameter.
func OfSlice[T any](elements []T) List[T] {
	result := Empty[T]()
	return result.AppendAll(elements)
}

// internal implementation of an non-empty List, consisting of a head of type T and a tail of type List[T].
type cons[T any] struct {
	consHead T
	consTail List[T]
	length   int
}

// newCons is an internal function used to create a new list
func newCons[T any](value T, tail List[T]) List[T] {
	return cons[T]{
		consHead: value,
		consTail: tail,
		length:   1 + tail.Length(),
	}
}

// head is an internal function used to get the head value of the current List
func (c cons[T]) head() T {
	return c.consHead
}

// tail is an internal function used to get the tail of the current List
func (c cons[T]) tail() List[T] {
	return c.consTail
}

// IsEmpty checks if the current List is empty.
// for the cons implementation of List interface, it always return false.
func (c cons[T]) IsEmpty() bool {
	return false
}

// Append returns a new List with the T value passed as parameter at the end of the new list created.
func (c cons[T]) Append(value T) List[T] {
	return newCons(c.consHead, c.consTail.Append(value))
}

// AppendAll returns a new List with the T values passed as an array at the end of the new list created.
func (c cons[T]) AppendAll(values []T) List[T] {
	result := newCons[T](c.consHead, c.consTail)
	for _, value := range values {
		result = result.Append(value)
	}
	return result
}

// Length returns the length of the current list.
func (c cons[T]) Length() int {
	return c.length
}

// Filter returns a new list containing only the elements which are validating the predicate passed as parameter.
func (c cons[T]) Filter(predicate func(T) bool) List[T] {
	if predicate(c.consHead) {
		return newCons[T](c.consHead, c.consTail.Filter(predicate))
	}
	if c.tail().IsEmpty() {
		return c.consTail
	}
	return newCons(c.consTail.head(), c.consTail.tail()).Filter(predicate)
}

// Remove returns a new list without all the elements matching the value passed as parameter.
func (c cons[T]) Remove(value T) List[T] {
	if !reflect.DeepEqual(c.consHead, value) {
		return newCons(c.consHead, c.consTail.Remove(value))
	}
	if c.consTail.IsEmpty() {
		return c.consTail
	}
	return newCons(c.consTail.head(), c.consTail.tail()).Remove(value)
}

// RemovePredicate returns a new list without all the elements matching the predicate passed as parameter.
func (c cons[T]) RemovePredicate(predicate func(T) bool) List[T] {
	if !predicate(c.consHead) {
		return newCons(c.consHead, c.consTail.RemovePredicate(predicate))
	}
	if c.consTail.IsEmpty() {
		return c.consTail
	}
	return newCons(c.consTail.head(), c.consTail.tail()).RemovePredicate(predicate)
}

// Insert returns a new list with the value passed as parameter at the position matching the index.
// this function returns error if the index is less than 0 or greater than list length.
func (c cons[T]) Insert(index int, value T) (List[T], error) {
	if index < 0 {
		return Empty[T](), fmt.Errorf("index out of range %d on List", index)
	}
	if index == 0 {
		return newCons[T](value, c), nil
	}
	if c.IsEmpty() {
		return Empty[T](), fmt.Errorf("index out of range %d on List", index)
	}
	tail, err := c.consTail.Insert(index-1, value)
	if err != nil {
		return Empty[T](), err
	}
	return newCons[T](c.consHead, tail), nil

}

// Reverse returns a reversed version of the current list.
func (c cons[T]) Reverse() List[T] {
	if c.IsEmpty() || c.length == 1 {
		return c
	}
	value := c.consHead
	list := c.consTail.Reverse()
	return list.Append(value)
}

// internal implementation of an empty list which could contain element of T type.
type empty[T any] struct{}

// head is an internal function used to get the head value of the current List
func (n empty[T]) head() T {
	return *new(T)
}

// tail is an internal function used to get the tail of the current List
func (n empty[T]) tail() List[T] {
	return Empty[T]()
}

// IsEmpty checks if the current List is empty.
// for the empty implementation of List interface, it always return true.
func (n empty[T]) IsEmpty() bool {
	return true
}

// Length returns the length of the current list.
func (n empty[T]) Length() int {
	return 0
}

// Append returns a new List with the T value passed as parameter at the end of the new list created.
// for the empty implementation of list interface, the function return a new single element List with value as head.
func (n empty[T]) Append(value T) List[T] {
	return newCons[T](value, Empty[T]())
}

// AppendAll returns a new List with the T values passed as an array at the end of the new list created.
func (n empty[T]) AppendAll(elements []T) List[T] {
	result := Empty[T]()
	for _, element := range elements {
		result = result.Append(element)
	}
	return result
}

// Filter returns a new list containing only the elements which are validating the predicate passed as parameter.
// for the empty implementation, the current empty list is returned.
func (n empty[T]) Filter(predicate func(T) bool) List[T] {
	return n
}

// Remove returns a new list without all the elements matching the value passed as parameter.
// for the empty implementation of the List interface, the current empty list is returned.
func (n empty[T]) Remove(value T) List[T] {
	return n
}

// RemovePredicate returns a new list without all the elements matching the predicate passed as parameter.
// for the empty implementation of the List interface, the current empty list is returned.
func (n empty[T]) RemovePredicate(predicate func(T) bool) List[T] {
	return n
}

// Insert returns a new list with the value passed as parameter at the position matching the index.
// this function returns error if the index is less than 0 or greater than list length.
func (n empty[T]) Insert(index int, value T) (List[T], error) {
	if index != 0 {
		return n, fmt.Errorf("index out of range %d on empty List", index)
	}
	return newCons[T](value, n), nil
}

// Reverse returns a reversed version of the current list.
// for the empty implementation, the current empty list is returned.
func (n empty[T]) Reverse() List[T] {
	return n
}

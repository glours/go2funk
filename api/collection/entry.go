package collection

import (
	"github.com/mitchellh/hashstructure/v2"
	"reflect"
)

type Entry[K comparable,V any] interface {
	GetKey() K
	GetValue() V
	Equals(entry Entry[K,V]) bool
	HashCode() (uint64, error)
}

type MapEntry[K comparable,V any] struct {
	key K
	value V
	left Entry[K,V]
	right Entry[K,V]
	parent Entry[K,V]
	color bool
}

func NewEntry[K comparable,V any](key K, value V, parent Entry[K,V]) Entry[K,V] {
	return MapEntry[K,V]{
		key: key,
		value: value,
		parent: parent,
	}
}

func (m MapEntry[K, V]) GetKey() K {
	return m.key
}

func (m MapEntry[K, V]) GetValue() V {
	return m.value
}

func (m MapEntry[K, V]) Equals(entry  Entry[K, V]) bool {
	return entry!= nil && entry.GetKey() == m.key && reflect.DeepEqual(entry.GetValue(),m.value)
}

func (m MapEntry[K, V]) HashCode() (uint64, error) {
	return hashstructure.Hash(m, hashstructure.FormatV2, nil)
}
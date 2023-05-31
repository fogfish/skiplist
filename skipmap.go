//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

package skiplist

type Map[K Key, V any] struct {
	keys   *Set[K]
	values map[K]V

	Length int
}

func NewMap[K Key, V any](opts ...ConfigSet[K]) *Map[K, V] {
	keys := NewSet(opts...)

	return &Map[K, V]{
		keys:   keys,
		values: make(map[K]V),
		Length: 0,
	}
}

func (kv *Map[K, V]) String() string { return kv.keys.String() }

func (kv *Map[K, V]) Put(key K, val V) bool {
	if _, has := kv.values[key]; has {
		kv.values[key] = val
		return false
	}

	kv.values[key] = val
	kv.keys.Add(key)
	kv.Length = kv.keys.Length
	return true
}

func (kv *Map[K, V]) Get(key K) (V, bool) {
	val, has := kv.values[key]
	return val, has
}

func (kv *Map[K, V]) Cut(key K) (V, bool) {
	val, has := kv.values[key]
	if has {
		delete(kv.values, key)
		kv.keys.Cut(key)
		kv.Length = kv.keys.Length
	}

	return val, has
}

func (kv *Map[K, V]) Keys() *Element[K] {
	return kv.keys.Values()
}

func (kv *Map[K, V]) Successors(key K) *Element[K] {
	return kv.keys.Successors(key)
}

// Predecessors elements from set
func (kv *Map[K, V]) Predecessor(key K) *Element[K] {
	return kv.keys.Predecessor(key)
}

func (kv *Map[K, V]) Neighbours(key K) (*Element[K], *Element[K]) {
	return kv.keys.Neighbours(key)
}

func (kv *Map[K, V]) Split(key K) *Map[K, V] {
	keys := kv.keys.Split(key)
	values := make(map[K]V)

	kv.Length = kv.keys.Length
	for e := keys.Values(); e != nil; e = e.Next() {
		values[e.key] = kv.values[e.key]
		delete(kv.values, e.key)
	}

	return &Map[K, V]{
		keys:   keys,
		values: values,
		Length: keys.Length,
	}
}

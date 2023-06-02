//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

package skiplist

type HashMap[K Key, V any] struct {
	keys   *Set[K]
	values map[K]V
}

func NewHashMap[K Key, V any](opts ...SetConfig[K]) *HashMap[K, V] {
	keys := NewSet(opts...)

	return &HashMap[K, V]{
		keys:   keys,
		values: make(map[K]V),
	}
}

func (kv *HashMap[K, V]) String() string { return kv.keys.String() }

func (set *HashMap[K, V]) Length() int { return set.keys.length }

func (set *HashMap[K, V]) Level() int { return set.keys.Level() }

func (kv *HashMap[K, V]) Put(key K, val V) bool {
	if _, has := kv.values[key]; has {
		kv.values[key] = val
		return false
	}

	kv.values[key] = val
	kv.keys.Add(key)
	return true
}

func (kv *HashMap[K, V]) Get(key K) (V, bool) {
	val, has := kv.values[key]
	return val, has
}

func (kv *HashMap[K, V]) Cut(key K) (V, bool) {
	val, has := kv.values[key]
	if has {
		delete(kv.values, key)
		kv.keys.Cut(key)
	}

	return val, has
}

func (kv *HashMap[K, V]) Keys() *Element[K] {
	return kv.keys.Values()
}

func (kv *HashMap[K, V]) Successor(key K) *Element[K] {
	return kv.keys.Successor(key)
}

func (kv *HashMap[K, V]) Predecessor(key K) *Element[K] {
	return kv.keys.Predecessor(key)
}

func (kv *HashMap[K, V]) Neighbours(key K) (*Element[K], *Element[K]) {
	return kv.keys.Neighbours(key)
}

func (kv *HashMap[K, V]) Split(key K) *HashMap[K, V] {
	keys := kv.keys.Split(key)
	values := make(map[K]V)

	for e := keys.Values(); e != nil; e = e.Next() {
		values[e.key] = kv.values[e.key]
		delete(kv.values, e.key)
	}

	return &HashMap[K, V]{
		keys:   keys,
		values: values,
	}
}

// --------------------------------------------------------------------------------------

// HashMapL[K] type projects HashMap[K] with all ops on level N
type HashMapL[K Key, V any] HashMap[K, V]

func ToHashMapL[K Key, V any](kv *HashMap[K, V]) *HashMapL[K, V] { return (*HashMapL[K, V])(kv) }

func (m *HashMapL[K, V]) Put(level int, key K, val V) bool {
	kv := (*HashMap[K, V])(m)

	if _, has := kv.values[key]; has {
		kv.values[key] = val
		return false
	}

	kv.values[key] = val
	ToSetL(kv.keys).Add(level, key)
	return true
}

// Cut segment on the level
func (m *HashMapL[K, V]) Cut(level int, from *Element[K]) map[K]V {
	kv := (*HashMap[K, V])(m)

	keys := ToSetL(kv.keys).Cut(level, from)
	values := make(map[K]V)

	for e := keys; e != nil; e = e.Next() {
		values[e.key] = kv.values[e.key]
		delete(kv.values, e.key)
	}

	return values
}

// All set elements on defined level
func (m *HashMapL[K, V]) Values(level int) *Element[K] {
	kv := (*HashMap[K, V])(m)
	return ToSetL(kv.keys).Values(level)
}

// Successor elements from set on given level
func (m *HashMapL[K, V]) Successor(level int, key K) *Element[K] {
	kv := (*HashMap[K, V])(m)
	return ToSetL(kv.keys).Successor(level, key)
}

func (m *HashMapL[K, V]) Predecessor(level int, key K) *Element[K] {
	kv := (*HashMap[K, V])(m)
	return kv.keys.Predecessor(key)
}

func (m *HashMapL[K, V]) Neighbours(level int, key K) (*Element[K], *Element[K]) {
	kv := (*HashMap[K, V])(m)
	return kv.keys.Neighbours(key)
}

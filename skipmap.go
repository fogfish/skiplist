//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

package skiplist

type SkipMap[K Key, V any] struct {
	keys   *Set[K]
	values map[K]V

	Length int
}

func NewMap[K Key, V any](opts ...ConfigSet[K]) *SkipMap[K, V] {
	keys := NewSet(opts...)

	return &SkipMap[K, V]{
		keys:   keys,
		values: make(map[K]V),
		Length: 0,
	}
}

func (kv *SkipMap[K, V]) String() string { return kv.keys.String() }

func (kv *SkipMap[K, V]) Put(key K, val V) bool {
	if _, has := kv.values[key]; has {
		kv.values[key] = val
		return false
	}

	kv.values[key] = val
	kv.keys.Add(key)
	kv.Length = kv.keys.Length
	return true
}

func (kv *SkipMap[K, V]) Get(key K) (V, bool) {
	val, has := kv.values[key]
	return val, has
}

func (kv *SkipMap[K, V]) Cut(key K) bool {
	delete(kv.values, key)
	flag := kv.keys.Cut(key)
	kv.Length = kv.keys.Length
	return flag
}

func (kv *SkipMap[K, V]) Keys() *Element[K] {
	return kv.keys.Values()
}

func (kv *SkipMap[K, V]) Successors(key K) *Element[K] {
	return kv.keys.Successors(key)
}

func (kv *SkipMap[K, V]) Split(key K) *SkipMap[K, V] {
	keys := kv.keys.Split(key)
	values := make(map[K]V)

	for e := keys.Values(); e != nil; e = e.Next() {
		values[e.key] = kv.values[e.key]
		delete(kv.values, e.key)
	}

	return &SkipMap[K, V]{
		keys:   keys,
		values: values,
	}
}

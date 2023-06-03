//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

package skiplist

import (
	"fmt"
	"strings"
)

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

func (kv *HashMap[K, V]) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("--- SkipHashMap[%T] %p ---\n", kv.keys.null, &kv))

	v := kv.keys.head
	for v != nil {
		sb.WriteString(v.String())
		sb.WriteString("\n")
		v = v.Fingers[0]
	}

	return sb.String()
}

func (kv *HashMap[K, V]) Length() int {
	return kv.keys.length
}

func (kv *HashMap[K, V]) Level() int {
	return kv.keys.Level()
}

func (kv *HashMap[K, V]) Skip(level int, key K) (*Element[K], [L]*Element[K]) {
	return kv.keys.Skip(level, key)
}

func (kv *HashMap[K, V]) Put(key K, val V) (bool, *Element[K]) {
	if _, has := kv.values[key]; has {
		kv.values[key] = val
		return false, nil
	}

	kv.values[key] = val
	return kv.keys.Add(key)
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

func (kv *HashMap[K, V]) Split(key K) *HashMap[K, V] {
	keys := kv.keys.Split(key)
	values := make(map[K]V)

	for e := keys.Values(); e != nil; e = e.Next() {
		values[e.Key] = kv.values[e.Key]
		delete(kv.values, e.Key)
	}

	return &HashMap[K, V]{
		keys:   keys,
		values: values,
	}
}

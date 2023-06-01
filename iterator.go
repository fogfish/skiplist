//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

package skiplist

import (
	"github.com/fogfish/golem/trait/pair"
	"github.com/fogfish/golem/trait/seq"
)

// Build generic iterate over Set elements
//
//	seq := skiplist.ForSet(set, set.Successor(key))
//	for has := seq != nil; has; has = seq.Next() {
//		seq.Key()
//	}
func ForSet[K Key](set *Set[K], el *Element[K]) seq.Seq[K] {
	if el == nil {
		return nil
	}
	return &forSet[K]{el}
}

type forSet[K Key] struct {
	el *Element[K]
}

func (it *forSet[K]) Value() K { return it.el.key }
func (it *forSet[K]) Next() bool {
	if it.el == nil {
		return false
	}

	it.el = it.el.Next()

	return it.el != nil
}

// Build generic iterate over Set elements on level N
//
//	seq := skiplist.ForSetOn(set, set.Values(...))
//	for has := seq != nil; has; has = seq.Next() {
//		seq.Key()
//	}
func ForSetOn[K Key](lvl int, set *Set[K], el *Element[K]) seq.Seq[K] {
	if el == nil {
		return nil
	}
	return &forSetOn[K]{lvl, el}
}

type forSetOn[K Key] struct {
	lvl int
	el  *Element[K]
}

func (it *forSetOn[K]) Value() K { return it.el.key }
func (it *forSetOn[K]) Next() bool {
	if it.el == nil {
		return false
	}

	it.el = it.el.NextOn(it.lvl)

	return it.el != nil
}

// Iterate over Map elements
//
//	seq := skiplist.ForMap(kv, kv.Successor(key))
//	for has := seq != nil; has; has = seq.Next() {
//		seq.Key()
//	}
func ForMap[K Key, V any](kv *Map[K, V], el *Pair[K, V]) pair.Seq[K, V] {
	if el == nil {
		return nil
	}

	return &forMap[K, V]{el: el}
}

type forMap[K Key, V any] struct {
	el *Pair[K, V]
}

func (it *forMap[K, V]) Key() K   { return it.el.key }
func (it *forMap[K, V]) Value() V { return it.el.value }
func (it *forMap[K, V]) Next() bool {
	if it.el == nil {
		return false
	}

	it.el = it.el.Next()

	return it.el != nil
}

func ForHashMap[K Key, V any](kv *HashMap[K, V], key *Element[K]) pair.Seq[K, V] {
	if key == nil {
		return nil
	}

	val, _ := kv.Get(key.key)
	return &forHashMap[K, V]{key: key, val: val, kv: kv}
}

func ForGF2[K Num](gf2 *GF2[K], key *Element[K]) pair.Seq[K, Arc[K]] {
	if key == nil {
		return nil
	}

	val, _ := gf2.Get(key.key)
	return &forHashMap[K, Arc[K]]{key: key, val: val, kv: gf2}
}

type getter[K Key, V any] interface {
	Get(K) (V, bool)
}

type forHashMap[K Key, V any] struct {
	key *Element[K]
	val V
	kv  getter[K, V]
}

func (it *forHashMap[K, V]) Key() K   { return it.key.key }
func (it *forHashMap[K, V]) Value() V { return it.val }
func (it *forHashMap[K, V]) Next() bool {
	if it.key == nil {
		return false
	}

	it.key = it.key.Next()
	if it.key == nil {
		return false
	}

	it.val, _ = it.kv.Get(it.key.key)

	return true
}

// Build generic iterate over Map elements on level N
//
//	seq := skiplist.ForMapOn(kv,  set.ValuesOn(...))
//	for has := seq != nil; has; has = seq.Next() {
//		seq.Key()
//	}
func ForMapOn[K Key, V any](lvl int, kv *Map[K, V], el *Pair[K, V]) pair.Seq[K, V] {
	if el == nil {
		return nil
	}
	return &forMapOn[K, V]{lvl, el}
}

type forMapOn[K Key, V any] struct {
	lvl int
	el  *Pair[K, V]
}

func (it *forMapOn[K, V]) Key() K   { return it.el.key }
func (it *forMapOn[K, V]) Value() V { return it.el.value }
func (it *forMapOn[K, V]) Next() bool {
	if it.el == nil {
		return false
	}

	it.el = it.el.NextOn(it.lvl)
	return it.el != nil
}

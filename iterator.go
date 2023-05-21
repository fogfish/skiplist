//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

package skiplist

// Generic iterator over skiplist data structures
// It is design to build operation over sequence of elements
//
//	seq := skiplist.ForSet(set, set.Successor(key))
//	for has := seq != nil; has; has = seq.Next() {
//		seq.Key()
//	}
type Iterator[K Key, V any] interface {
	Key() K
	Value() V
	Next() bool
}

// Iterate over Set elements
//
//	seq := skiplist.ForSet(set, set.Successor(key))
//	for has := seq != nil; has; has = seq.Next() {
//		seq.Key()
//	}
func ForSet[K Key](set *Set[K], el *Element[K]) Iterator[K, K] {
	if el == nil {
		return nil
	}
	return &forSet[K]{el}
}

type forSet[K Key] struct {
	el *Element[K]
}

func (it *forSet[K]) Key() K   { return it.el.key }
func (it *forSet[K]) Value() K { return it.el.key }
func (it *forSet[K]) Next() bool {
	if it.el == nil {
		return false
	}

	it.el = it.el.Next()

	return it.el != nil
}

type Getter[K Key, V any] interface {
	Get(K) (V, bool)
}

// Iterate over Map elements
//
//	seq := skiplist.ForMap(kv, kv.Successor(key))
//	for has := seq != nil; has; has = seq.Next() {
//		seq.Key()
//	}
func ForMap[K Key, V any](kv *Map[K, V], key *Element[K]) Iterator[K, V] {
	if key == nil {
		return nil
	}

	val, _ := kv.Get(key.key)
	return &forMap[K, V]{key: key, val: val, kv: kv}
}

func ForGF2[K Num](gf2 *GF2[K], key *Element[K]) Iterator[K, Arc[K]] {
	if key == nil {
		return nil
	}

	val, _ := gf2.Get(key.key)
	return &forMap[K, Arc[K]]{key: key, val: val, kv: gf2}
}

type forMap[K Key, V any] struct {
	key *Element[K]
	val V
	kv  Getter[K, V]
}

func (it *forMap[K, V]) Key() K   { return it.key.key }
func (it *forMap[K, V]) Value() V { return it.val }
func (it *forMap[K, V]) Next() bool {
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

// Take values from iterator while predicate function true
func TakeWhile[K Key, V any](seq Iterator[K, V], f func(K, V) bool) Iterator[K, V] {
	if seq == nil || !f(seq.Key(), seq.Value()) {
		return nil
	}

	return &takeWhile[K, V]{
		Iterator: seq,
		f:        f,
	}
}

type takeWhile[K Key, V any] struct {
	Iterator[K, V]
	f func(K, V) bool
}

func (seq *takeWhile[K, V]) Next() bool {
	if seq.f == nil || seq.Iterator == nil {
		return false
	}

	if !seq.Iterator.Next() {
		return false
	}

	if !seq.f(seq.Key(), seq.Value()) {
		seq.f = nil
		return false
	}

	return true
}

// Drop values from iterator while predicate function true
func DropWhile[K Key, V any](seq Iterator[K, V], f func(K, V) bool) Iterator[K, V] {
	for {
		if !f(seq.Key(), seq.Value()) {
			return seq
		}

		if !seq.Next() {
			return nil
		}
	}
}

// Filter values from iterator
func Filter[K Key, V any](seq Iterator[K, V], f func(K, V) bool) Iterator[K, V] {
	for {
		if f(seq.Key(), seq.Value()) {
			return filter[K, V]{
				Iterator: seq,
				f:        f,
			}
		}

		if !seq.Next() {
			return nil
		}
	}
}

type filter[K Key, V any] struct {
	Iterator[K, V]
	f func(K, V) bool
}

func (seq filter[K, V]) Next() bool {
	if seq.f == nil || seq.Iterator == nil {
		return false
	}

	for {
		if !seq.Iterator.Next() {
			return false
		}

		if seq.f(seq.Key(), seq.Value()) {
			return true
		}
	}
}

// ForEach applies clojure on iterator
func ForEach[K Key, V any](seq Iterator[K, V], f func(K, V) error) error {
	for has := seq != nil; has; has = seq.Next() {
		if err := f(seq.Key(), seq.Value()); err != nil {
			return err
		}
	}

	return nil
}

// FMap transform iterator type
func FMap[K Key, A, B any](seq Iterator[K, A], f func(K, A) B) Iterator[K, B] {
	return fmap[K, A, B]{Iterator: seq, f: f}
}

type fmap[K Key, A, B any] struct {
	Iterator[K, A]
	f func(K, A) B
}

func (seq fmap[K, A, B]) Value() B {
	return seq.f(seq.Iterator.Key(), seq.Iterator.Value())
}

// Plus operation for iterators add one after another
func Plus[K Key, V any](lhs, rhs Iterator[K, V]) Iterator[K, V] {
	if lhs == nil {
		return rhs
	}

	if rhs == nil {
		return lhs
	}

	return &plus[K, V]{Iterator: lhs, rhs: rhs}
}

type plus[K Key, V any] struct {
	Iterator[K, V]
	rhs Iterator[K, V]
}

func (plus *plus[K, V]) Next() bool {
	hasNext := plus.Iterator.Next()

	if !hasNext && plus.rhs != nil {
		plus.Iterator, plus.rhs = plus.rhs, nil
		return true
	}

	if !hasNext && plus.rhs == nil {
		return false
	}

	return true
}

// Left join
func Join[K1, K2 Key, V1, V2 any](lhs Iterator[K1, V1], rhs func(K1, V1) Iterator[K2, V2]) Iterator[K2, V2] {
	if lhs == nil {
		return nil
	}

	join := &join[K1, K2, V1, V2]{lhs: lhs, rhs: rhs}
	for {
		join.Iterator = join.rhs(join.lhs.Key(), join.lhs.Value())
		if join.Iterator != nil {
			return join
		}

		if !join.lhs.Next() {
			return nil
		}
	}
}

type join[K1, K2 Key, V1, V2 any] struct {
	Iterator[K2, V2]
	lhs Iterator[K1, V1]
	rhs func(K1, V1) Iterator[K2, V2]
}

func (join *join[K1, K2, V1, V2]) Next() bool {
	if !join.Iterator.Next() {
		for {
			if !join.lhs.Next() {
				return false
			}

			join.Iterator = join.rhs(join.lhs.Key(), join.lhs.Value())
			if join.Iterator != nil {
				return true
			}
		}
	}

	return true
}

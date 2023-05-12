//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

package skiplist

// Iterator over Skip List nodes.
// Newly created iterator holds the first element in sequence,
// consume it with KeyValue(), forward with Next()
//
//	for has := seq != nil; has; has = seq.Next() {
//		seq.KeyValue()
//	}
type Iterator[K, V any] interface {
	Key() K
	Value() V
	KeyValue() (K, V)
	Next() bool
}

type iterator[K, V any] struct {
	*Node[K, V]
}

func newIterator[K, V any](node *Node[K, V]) Iterator[K, V] {
	return &iterator[K, V]{Node: node}
}

// Next element
func (seq *iterator[K, V]) Next() bool {
	seq.Node = seq.Node.fingers[0]
	return seq.Node != nil
}

// Take values from iterator while predicate function true
func TakeWhile[K, V any](seq Iterator[K, V], f func(K, V) bool) Iterator[K, V] {
	if seq == nil || !f(seq.KeyValue()) {
		return nil
	}

	return &takeWhile[K, V]{
		Iterator: seq,
		f:        f,
	}
}

type takeWhile[K, V any] struct {
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

	if !seq.f(seq.KeyValue()) {
		seq.f = nil
		return false
	}

	return true
}

// Drop values from iterator while predicate function true
func DropWhile[K, V any](seq Iterator[K, V], f func(K, V) bool) Iterator[K, V] {
	for {
		if !f(seq.KeyValue()) {
			return seq
		}

		if !seq.Next() {
			return nil
		}
	}
}

// Filter values from iterator
func Filter[K, V any](seq Iterator[K, V], f func(K, V) bool) Iterator[K, V] {
	for {
		if f(seq.KeyValue()) {
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

type filter[K, V any] struct {
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

		if seq.f(seq.KeyValue()) {
			return true
		}
	}
}

// FMap applies clojure on iterator
func FMap[K, V any](seq Iterator[K, V], f func(K, V) error) error {
	for seq.Next() {
		if err := f(seq.KeyValue()); err != nil {
			return err
		}
	}

	return nil
}

// Map transform iterator type
func Map[K, A, B any](seq Iterator[K, A], f func(K, A) B) Iterator[K, B] {
	return mapper[K, A, B]{Iterator: seq, f: f}
}

type mapper[K, A, B any] struct {
	Iterator[K, A]
	f func(K, A) B
}

func (seq mapper[K, A, B]) Value() B {
	return seq.f(seq.Iterator.KeyValue())
}

func (seq mapper[K, A, B]) KeyValue() (K, B) {
	return seq.Iterator.Key(), seq.f(seq.Iterator.KeyValue())
}

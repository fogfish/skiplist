//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

package skiplist

import "github.com/fogfish/skiplist/ord"

type Iterator[K, V any] struct {
	ord         ord.Ord[K]
	node, until *tSkipNode[K, V]
}

func newIterator[K, V any](ord ord.Ord[K], node, until *tSkipNode[K, V]) *Iterator[K, V] {
	return &Iterator[K, V]{
		ord:   ord,
		node:  node,
		until: until,
	}
}

// Head element of the iterator
func (seq *Iterator[K, V]) Head() (K, V) {
	return seq.node.key, seq.node.val
}

// Next element
func (seq *Iterator[K, V]) Next() bool {
	seq.node = seq.node.fingers[0]
	if seq.until == nil {
		return seq.node != nil
	}

	return seq.node != nil && seq.ord.Compare(seq.node.key, seq.until.key) == -1
}

// FMap applies clojure on iterator
func (seq *Iterator[K, V]) FMap(f func(K, V) error) error {
	for seq.Next() {
		if err := f(seq.Head()); err != nil {
			return err
		}
	}

	return nil
}

//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

package skiplist

import "github.com/fogfish/skiplist/ord"

type Seq[K, V any] struct {
	ord         ord.Ord[K]
	node, until *tSkipNode[K, V]
}

func newSeq[K, V any](ord ord.Ord[K], node, until *tSkipNode[K, V]) *Seq[K, V] {
	return &Seq[K, V]{
		ord:   ord,
		node:  node,
		until: until,
	}
}

func (seq *Seq[K, V]) Head() (K, V) {
	return seq.node.key, seq.node.val
}

func (seq *Seq[K, V]) Tail() bool {
	seq.node = seq.node.fingers[0]
	if seq.until == nil {
		return seq.node != nil
	}

	return seq.ord.Compare(seq.node.key, seq.until.key) == -1
}

func (seq *Seq[K, V]) FMap(f func(K, V) error) error {
	for seq.Tail() {
		if err := f(seq.Head()); err != nil {
			return err
		}
	}

	return nil
}

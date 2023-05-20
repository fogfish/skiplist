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
	"reflect"
	"strings"
)

type Ftype interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Field[K Ftype] struct {
	keys *Set[K]
	arcs map[K]arc[K]
}

type arc[K Ftype] struct {
	rank uint32
	lo   K
}

func NewField[K Ftype](opts ...ConfigSet[K]) *Field[K] {
	keys := NewSet(opts...)

	top := *new(K) - 1
	keys.Add(top)
	rnk := uint32(reflect.TypeOf(top).Size() * 8)

	arcs := map[K]arc[K]{
		top: {rank: rnk, lo: 0},
	}

	return &Field[K]{
		keys: keys,
		arcs: arcs,
	}
}

func (f *Field[K]) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("--- SkipField[%T] %p ---\n", *new(K), &f))

	for node := f.keys.Values(); node != nil; node = node.Next() {
		key := node.Key()
		arc := f.arcs[key]
		sb.WriteString(
			fmt.Sprintf("{ %2d : %8x - %8x | %10d - %10d }\n", arc.rank, arc.lo, key, arc.lo, key),
		)
	}

	return sb.String()
}

func (f *Field[K]) Where(key K) (K, K) {
	node := f.keys.Successors(key)
	if node == nil {
		panic("non-continuos field")
	}

	hi := node.key
	arc := f.arcs[hi]

	return arc.lo, hi
}

func (f *Field[K]) Split(key K) (K, K, K) {
	node := f.keys.Successors(key)
	if node == nil {
		panic("non-continuos field")
	}

	hi := node.key
	tail := f.arcs[hi]

	if tail.rank == 0 {
		return tail.lo, hi, hi
	}

	rnk := tail.rank - 1
	mid := tail.lo + (hi-tail.lo)/2

	head := arc[K]{rank: rnk, lo: tail.lo}
	tail.rank, tail.lo = rnk, mid+1

	f.keys.Add(mid)
	f.arcs[mid] = head
	f.arcs[hi] = tail

	return head.lo, mid, hi
}

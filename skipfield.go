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

type Num interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

type GF2[K Num] struct {
	keys *Set[K]
	arcs map[K]Arc[K]
}

type Arc[K Num] struct {
	Rank   uint32
	Lo, Hi K
}

func (arc Arc[K]) String() string {
	return fmt.Sprintf("{ %2d : %8x - %8x | %10d - %10d }", arc.Rank, arc.Lo, arc.Hi, arc.Lo, arc.Hi)
}

func NewGF2[K Num](opts ...SetConfig[K]) *GF2[K] {
	keys := NewSet(opts...)

	top := *new(K) - 1
	keys.Add(top)
	rnk := uint32(reflect.TypeOf(top).Size() * 8)

	arcs := map[K]Arc[K]{
		top: {Rank: rnk, Lo: 0, Hi: top},
	}

	return &GF2[K]{
		keys: keys,
		arcs: arcs,
	}
}

func (f *GF2[K]) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("--- SkipGF2[%T] %p ---\n", *new(K), &f))

	for node := f.keys.Values(); node != nil; node = node.Next() {
		key := node.Key
		arc := f.arcs[key]
		sb.WriteString(arc.String())
		sb.WriteString("\n")
	}

	return sb.String()
}

func (f *GF2[K]) Length() int { return f.keys.length }

// Add new element to the field
func (f *GF2[K]) Add(key K) (Arc[K], Arc[K]) {
	node := f.keys.Successor(key)
	if node == nil {
		panic("non-continuos field")
	}

	hi := node.Key
	tail := f.arcs[hi]

	if tail.Rank == 0 {
		return tail, tail
	}

	rnk := tail.Rank - 1
	mid := tail.Lo + (hi-tail.Lo)/2

	head := Arc[K]{Rank: rnk, Lo: tail.Lo, Hi: mid}
	tail.Rank, tail.Lo = rnk, mid+1

	f.keys.Add(mid)
	f.arcs[mid] = head
	f.arcs[hi] = tail

	return head, tail
}

// Put element
func (f *GF2[K]) Put(arc Arc[K]) bool {
	added, _ := f.keys.Add(arc.Hi)

	f.arcs[arc.Hi] = arc

	return added
}

// Check elements position on the field
func (f *GF2[K]) Get(key K) (Arc[K], bool) {
	node := f.keys.Successor(key)
	if node == nil {
		panic("non-continuos field")
	}

	return f.arcs[node.Key], true
}

func (f *GF2[K]) Keys() *Element[K] {
	return f.keys.Values()
}

func (f *GF2[K]) Successor(key K) *Element[K] {
	return f.keys.Successor(key)
}

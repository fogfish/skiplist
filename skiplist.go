//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

/*

Package skiplist implements a probabilistic list-based data structure
that are a simple and efficient substitute for balanced trees.

Please see the article that depicts the data structure
http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.17.524
*/
package skiplist

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/fogfish/skiplist/ord"
)

/*

L depth of fingers at each node.

The value is estimated as math.Log10(float64(n)) / math.Log10(1/p)
n = 4294967296, p = 1/math.E
*/
const L = 22

/*

SkipList data structure
*/
type SkipList[K, V any] struct {
	ord ord.Ord[K]

	//
	// head of the list, the node is a lowest element
	head *tSkipNode[K, V]

	//
	// number of elements in the list, O(1)
	length int

	//
	// random generator
	random rand.Source

	//
	// probability table to determine node level
	p []float64

	//
	// buffer to estimate the skip path during insert / remove
	// the buffer implements optimization of memory allocations
	path [L]*tSkipNode[K, V]
}

// String converts table to string
func (list *SkipList[K, V]) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("--- SkipList %p ---\n", &list))

	v := list.head
	for v != nil {
		buffer.WriteString(v.String())
		buffer.WriteString("\n")
		v = v.fingers[0]
	}

	return buffer.String()
}

/*

New create instance of SkipList
*/
func New[K, V any](ord ord.Ord[K]) *SkipList[K, V] {
	ptable := probability(4294967296, 1/math.E)
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	return &SkipList[K, V]{
		ord:    ord,
		head:   newSkipNode[K, V](L),
		length: 0,
		random: random,
		p:      ptable,
		path:   [L]*tSkipNode[K, V]{},
	}
}

/*

calculates probability table
*/
func probability(n int, p float64) []float64 {
	// level := int(math.Log10(float64(n)) / math.Log10(1/p))
	table := make([]float64, L+1)

	for i := 1; i <= L; i++ {
		table[i-1] = math.Pow(p, float64(i-1))
	}

	return /*level,*/ table
}

/*

Length number of elements in data structure
*/
func Length[K, V any](list *SkipList[K, V]) int {
	return list.length
}

/*

Put insert the element into the list
*/
func Put[K, V any](list *SkipList[K, V], key K, val V) *SkipList[K, V] {
	v, path := skip(list, key)

	if v != nil && list.ord.Compare(v.key, key) == 0 {
		v.val = val
		return list
	}

	rank, node := mkNode(list, key, val)

	// re-bind fingers to new node
	for level := 0; level < rank; level++ {
		node.fingers[level] = path[level].fingers[level]
		path[level].fingers[level] = node
	}

	list.length++
	return list
}

/*

skip algorithm is similar to search algorithm that traversing forward pointers.
skip maintain the vector path that contains a pointer to the rightmost node
of level i or higher that is to the left of the location of the
insertion/deletion.
*/
func skip[K, V any](list *SkipList[K, V], key K) (*tSkipNode[K, V], [L]*tSkipNode[K, V]) {
	path := list.path

	node := list.head
	next := node.fingers
	for level := L - 1; level >= 0; level-- {
		for next[level] != nil && list.ord.Compare(next[level].key, key) == -1 {
			node = node.fingers[level]
			next = node.fingers
		}
		path[level] = node
	}

	return next[0], path
}

/*

mkNode creates a new node, randomly defines empty fingers (level of the node)
*/
func mkNode[K, V any](list *SkipList[K, V], key K, val V) (int, *tSkipNode[K, V]) {
	// See: https://golang.org/src/math/rand/rand.go#L150
	p := float64(list.random.Int63()) / (1 << 63)

	level := 0
	for level < L && p < list.p[level] {
		level++
	}

	node := &tSkipNode[K, V]{
		key:     key,
		val:     val,
		fingers: [L]*tSkipNode[K, V]{},
	}

	return level, node
}

/*

Get looks up the element in the list
*/
func Get[K, V any](list *SkipList[K, V], key K) V {
	if v, has := Lookup(list, key); has {
		return v
	}

	return *new(V)
}

/*

Lookup the element in the list, return bool flag
*/
func Lookup[K, V any](list *SkipList[K, V], key K) (V, bool) {
	node := search(list, key)

	if node != nil && list.ord.Compare(node.key, key) == 0 {
		return node.val, true
	}

	return *new(V), false
}

/*

search algorithm traversing forward pointers that do not jumps over the node
containing the element (for each level the finger shall be less than key).
When no more progress can be made at the current level of forward pointers,
the search moves down to the next level. When we can make no more progress at
level 0, we must be immediately in front of the node that contains
the desired element (if it is in the list).
*/
func search[K, V any](list *SkipList[K, V], key K) *tSkipNode[K, V] {
	node := list.head
	next := list.head.fingers
	for level := L - 1; level >= 0; level-- {
		for next[level] != nil && list.ord.Compare(next[level].key, key) == -1 {
			node = node.fingers[level]
			next = node.fingers
		}
	}

	return next[0]
}

/*

Remove element from the list
*/
func Remove[K, V any](list *SkipList[K, V], key K) V {
	rank := len(list.head.fingers)
	v, path := skip(list, key)

	if v != nil && list.ord.Compare(v.key, key) == 0 {
		for level := 0; level < rank; level++ {
			if path[level].fingers[level] == v {
				if len(v.fingers) > level {
					path[level].fingers[level] = v.fingers[level]
				} else {
					path[level].fingers[level] = nil
				}
			}
		}
		list.length--
		return v.val
	}

	return *new(V)
}

/*

Values return all values from the list
*/
func Values[K, V any](list *SkipList[K, V]) *Iterator[K, V] {
	return newIterator(list.ord, list.head, nil)
}

/*

Split the list
*/
func Split[K, V any](list *SkipList[K, V], key K) (*Iterator[K, V], *Iterator[K, V]) {
	v, p := skip(list, key)

	head := newIterator(list.ord, p[L-1], v)
	tail := newIterator(list.ord, p[0], nil)

	if v == nil {
		return head, nil
	}

	if p[0] == p[L-1] && list.ord.Compare(v.key, key) != 0 {
		return nil, tail
	}

	return head, tail
}

/*

Range iterates the list on the inclusive range [from, to]
*/
func Range[K, V any](list *SkipList[K, V], from, to K) *Iterator[K, V] {
	v, p := skip(list, from)

	if v == nil {
		return nil
	}

	iter := newIterator[K](inclusiveRange[K]{list.ord}, p[0], &tSkipNode[K, V]{key: to})
	return iter
}

type inclusiveRange[K any] struct{ ord.Ord[K] }

func (inclusive inclusiveRange[K]) Compare(a, b K) int {
	if inclusive.Ord.Compare(a, b) != 1 {
		return -1
	}
	return 1
}

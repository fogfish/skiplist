//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

// Package skiplist implements a probabilistic list-based data structure
// that are a simple and efficient substitute for balanced trees.
//
// Please see the article that depicts the data structure
// https://15721.courses.cs.cmu.edu/spring2018/papers/08-oltpindexes1/pugh-skiplists-cacm1990.pdf
// http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.17.524
package skiplist

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	"github.com/fogfish/skiplist/ord"
)

// L depth of fingers at each node.
//
// The value is estimated as math.Log10(float64(n)) / math.Log10(1/p)
// n = 4294967296, p = 1/math.E
const L = 22

// The probability table is generated for L=22
var probabilityTable []float64 = []float64{1, 0.36787944117144233, 0.1353352832366127, 0.04978706836786395, 0.018315638888734182, 0.006737946999085468, 0.002478752176666359, 0.0009118819655545165, 0.0003354626279025119, 0.0001234098040866796, 4.539992976248486e-05, 1.6701700790245666e-05, 6.1442123533282115e-06, 2.260329406981055e-06, 8.315287191035682e-07, 3.0590232050182594e-07, 1.1253517471925916e-07, 4.139937718785168e-08, 1.5229979744712636e-08, 5.60279643753727e-09, 2.0611536224385587e-09, 7.582560427911911e-10, 0}

// SkipList is a probabilistic list-based data structure
type SkipList[K, V any] struct {
	ord ord.Ord[K]

	//
	// head of the list, the node is a lowest element
	head *Node[K, V]

	//
	// number of elements in the list, O(1)
	length int

	//
	// random generator
	random rand.Source

	//
	// buffer to estimate the skip path during insert / remove
	// the buffer implements optimization of memory allocations
	path [L]*Node[K, V]
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

// Head (first) element of the list
func (list *SkipList[K, V]) Head() *Node[K, V] { return list.head.fingers[0] }

// New create instance of SkipList
func New[K, V any](ord ord.Ord[K], random ...rand.Source) *SkipList[K, V] {
	// ptable := probability(1<<32, 1/math.E)

	var rnd rand.Source
	if len(random) == 0 {
		rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	} else {
		rnd = random[0]
	}

	return &SkipList[K, V]{
		ord:    ord,
		head:   newNode[K, V](L),
		length: 0,
		random: rnd,
		path:   [L]*Node[K, V]{},
	}
}

// calculates probability table
// func probability(n int, p float64) []float64 {
// 	// level := int(math.Log10(float64(n)) / math.Log10(1/p))
// 	table := make([]float64, L+1)
//
// 	for i := 1; i <= L; i++ {
// 		table[i-1] = math.Pow(p, float64(i-1))
// 	}
//
// 	return /*level,*/ table
// }

// Length number of elements in data structure
func Length[K, V any](list *SkipList[K, V]) int {
	return list.length
}

// Put insert the element into the list
func Put[K, V any](list *SkipList[K, V], key K, val V) {
	v, path := skip(list, key)

	if v != nil && list.ord.Compare(v.key, key) == 0 {
		v.val = val
		return
	}

	rank, node := mkNode(list, key, val)

	// re-bind fingers to new node
	for level := 0; level < rank; level++ {
		node.fingers[level] = path[level].fingers[level]
		path[level].fingers[level] = node
	}

	list.length++
}

// skip algorithm is similar to search algorithm that traversing forward pointers.
// skip maintain the vector path that contains a pointer to the rightmost node
// of level i or higher that is to the left of the location of the
// insertion/deletion.
func skip[K, V any](list *SkipList[K, V], key K) (*Node[K, V], [L]*Node[K, V]) {
	path := list.path

	node := list.head
	next := &node.fingers
	for level := L - 1; level >= 0; level-- {
		for next[level] != nil && list.ord.Compare(next[level].key, key) == -1 {
			node = node.fingers[level]
			next = &node.fingers
		}
		path[level] = node
	}

	return next[0], path
}

// mkNode creates a new node, randomly defines empty fingers (level of the node)
func mkNode[K, V any](list *SkipList[K, V], key K, val V) (int, *Node[K, V]) {
	// See: https://golang.org/src/math/rand/rand.go#L150
	p := float64(list.random.Int63()) / (1 << 63)

	level := 0
	for level < L && p < probabilityTable[level] {
		level++
	}

	node := &Node[K, V]{
		key:     key,
		val:     val,
		fingers: [L]*Node[K, V]{},
	}

	return level, node
}

// Get looks up the element in the list.
// It return "empty" value for type V is key is not found
// TODO: Get(...) (V, bool)
func Get[K, V any](list *SkipList[K, V], key K) V {
	if node := Lookup(list, key); node != nil {
		return node.val
	}

	return *new(V)
}

// Lookup node equal to key.
// It return nil if element is not found
func Lookup[K, V any](list *SkipList[K, V], key K) *Node[K, V] {
	node := search(list, key)

	if node != nil && list.ord.Compare(node.key, key) == 0 {
		return node
	}

	return nil
}

// Lookup first node either equal of greater than key [key, ...)
func LookupAfter[K, V any](list *SkipList[K, V], key K) *Node[K, V] {
	node := search(list, key)
	// node, _ := skip(list, key)

	if node == list.head {
		return list.head.fingers[0]
	}

	return node
}

// Lookup first node either equal of lower than key (..., key]
func LookupBefore[K, V any](list *SkipList[K, V], key K) *Node[K, V] {
	node, path := skip(list, key)

	if node != nil && list.ord.Compare(node.key, key) == 0 {
		return node
	}

	if path[0] == list.head {
		return list.head.fingers[0]
	}

	return path[0]
}

// search algorithm traversing forward pointers that do not jumps over the node
// containing the element (for each level the finger shall be less than key).
// When no more progress can be made at the current level of forward pointers,
// the search moves down to the next level. When we can make no more progress at
// level 0, we must be immediately in front of the node that contains
// the desired element (if it is in the list).
func search[K, V any](list *SkipList[K, V], key K) *Node[K, V] {
	node := list.head
	next := &list.head.fingers
	for level := L - 1; level >= 0; level-- {
		for next[level] != nil && list.ord.Compare(next[level].key, key) == -1 {
			node = node.fingers[level]
			next = &node.fingers
		}
	}

	return next[0]
}

// Remove element from the list
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

// Values return all values from the list
func Values[K, V any](list *SkipList[K, V]) Iterator[K, V] {
	return newIterator(list.Head())
}

// Split the list before and after the key.
// It returns two lists [..., key) and [key, ...].
func Split[K, V any](list *SkipList[K, V], key K) (Iterator[K, V], Iterator[K, V]) {
	v, p := skip(list, key)

	if v == nil {
		return list.Head().Seq(), nil
	}

	if p[0] == p[L-1] {
		return nil, list.Head().Seq()
	}

	head := TakeWhile(list.Head().Seq(),
		func(k K, v V) bool { return list.ord.Compare(k, key) == -1 },
	)
	tail := newIterator(v)

	return head, tail
}

// Split the list at key into distinct lists [..., key) and [key, ...].
func SplitAt[K, V any](list *SkipList[K, V], key K) (*SkipList[K, V], *SkipList[K, V]) {
	v, p := skip(list, key)

	for level, x := range p {
		x.fingers[level] = nil
	}

	count := 0
	for node := list.head.fingers[0]; node != nil; node = node.fingers[0] {
		count++
	}

	tail := &SkipList[K, V]{
		ord:    list.ord,
		head:   newNode[K, V](L),
		length: list.length - count,
		random: list.random,
		path:   [L]*Node[K, V]{},
	}
	tail.head.fingers[0] = v

	list.length = count

	return list, tail
}

// Extract inclusive range of the list on [from, to] interval
func Range[K, V any](list *SkipList[K, V], from, to K) Iterator[K, V] {
	v, _ := skip(list, from)

	if v == nil {
		return nil
	}

	return TakeWhile(newIterator(v),
		func(k K, v V) bool { return list.ord.Compare(k, to) != 1 },
	)
}

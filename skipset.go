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
	"math/rand"
	"strings"
	"time"
)

// Abstract element of set
type Element[K Key] struct {
	key     K
	fingers [L]*Element[K]
}

// Value of element
func (el *Element[K]) Key() K { return el.key }

// Return next element in the set.
// Use for-loop to iterate through set elements
//
//	for e := set.Successor(...); e != nil; e.Next() { /* ... */}
func (el *Element[K]) Next() *Element[K] { return el.fingers[0] }

// Cast Element into string
func (el *Element[K]) String() string {
	fingers := ""
	for _, x := range el.fingers {
		if x != nil {
			fingers = fingers + fmt.Sprintf(" %v", x.key)
		}
	}

	return fmt.Sprintf("{ %4v\t|%s }", el.key, fingers)
}

// Configure Set properties
type ConfigSet[K Key] func(*Set[K])

// Configure Random Generator
func ConfigSetRandomSource[K Key](random rand.Source) ConfigSet[K] {
	return func(set *Set[K]) {
		set.random = random
	}
}

// Configure Memory Allocator
func ConfigSetAllocator[K Key](malloc Allocator[K, Element[K]]) ConfigSet[K] {
	return func(set *Set[K]) {
		set.malloc = malloc
	}
}

// Set of Elements
type Set[K Key] struct {
	//
	// head of the list, the node is a lowest element
	head *Element[K]

	// null element of type T
	null K

	//
	// number of elements in the set, O(1)
	Length int

	//
	// random generator
	random rand.Source

	//
	// buffer to estimate the skip path during insert / remove
	// the buffer implements optimization of memory allocations
	path [L]*Element[K]

	// memory allocator for elements
	malloc Allocator[K, Element[K]]
}

// New create instance of SkipList
func NewSet[K Key](opts ...ConfigSet[K]) *Set[K] {
	set := &Set[K]{
		head:   new(Element[K]),
		null:   *new(K),
		Length: 0,
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
		path:   [L]*Element[K]{},
		malloc: malloc[K, Element[K]]{},
	}

	for _, opt := range opts {
		opt(set)
	}

	return set
}

// Cast set into string
func (set *Set[K]) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("--- SkipSet[%T] %p ---\n", set.null, &set))

	v := set.head
	for v != nil {
		sb.WriteString(v.String())
		sb.WriteString("\n")
		v = v.fingers[0]
	}

	return sb.String()
}

// skip algorithm is similar to search algorithm that traversing forward pointers.
// skip maintain the vector path that contains a pointer to the rightmost node
// of level i or higher that is to the left of the location of the
// insertion/deletion.
func (set *Set[K]) skip(key K) (*Element[K], [L]*Element[K]) {
	path := set.path

	node := set.head
	next := &node.fingers
	for level := L - 1; level >= 0; level-- {
		for next[level] != nil && next[level].key < key {
			node = node.fingers[level]
			next = &node.fingers
		}
		path[level] = node
	}

	return next[0], path
}

// Add element to set, return true if element is new
func (set *Set[K]) Add(key K) bool {
	el, path := set.skip(key)

	if el != nil && el.key == key {
		return false
	}

	rank, el := set.createElement(key)

	// re-bind fingers to new node
	for level := 0; level < rank; level++ {
		el.fingers[level] = path[level].fingers[level]
		path[level].fingers[level] = el
	}

	set.Length++
	return true
}

// mkNode creates a new node, randomly defines empty fingers (level of the node)
func (set *Set[K]) createElement(key K) (int, *Element[K]) {
	// See: https://golang.org/src/math/rand/rand.go#L150
	p := float64(set.random.Int63()) / (1 << 63)

	level := 0
	for level < L && p < probabilityTable[level] {
		level++
	}

	var node *Element[K]
	if set.malloc == nil {
		node = new(Element[K])
	} else {
		node = set.malloc.Alloc(key)
	}
	node.key = key

	return level, node
}

// Check is element exists in set
func (set *Set[K]) Has(key K) bool {
	el, _ := set.skip(key)

	if el != nil && el.key == key {
		return true
	}

	return false
}

// Cut element from the set, returns true if element is removed
func (set *Set[K]) Cut(key K) bool {
	rank := L
	v, path := set.skip(key)

	if v == nil || v.key != key {
		return false
	}

	for level := 0; level < rank; level++ {
		if path[level].fingers[level] == v {
			if len(v.fingers) > level {
				path[level].fingers[level] = v.fingers[level]
			} else {
				path[level].fingers[level] = nil
			}
		}
	}

	set.Length--

	if set.malloc != nil {
		set.malloc.Free(key)
	}

	return true
}

// All set elements
func (set *Set[K]) Values() *Element[K] {
	return set.head.fingers[0]
}

// Successor elements from set
func (set *Set[K]) Successors(key K) *Element[K] {
	el, _ := set.skip(key)
	return el
}

// Split set of elements by key
func (set *Set[K]) Split(key K) *Set[K] {
	node, path := set.skip(key)

	for level, x := range path {
		x.fingers[level] = nil
	}

	tail := &Set[K]{
		head:   new(Element[K]),
		null:   *new(K),
		Length: 0,
		random: set.random,
		path:   [L]*Element[K]{},
		malloc: set.malloc,
	}
	tail.head.fingers[0] = node

	length := 0
	for n := node; n != nil; n = n.fingers[0] {
		length++
	}

	tail.Length = length
	set.Length -= length

	return tail
}

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
	"math"
	"math/rand"
	"strings"
	"time"
)

// Each element is represented by a Element in a skip structures. Each node has
// a height or level (length of fingers array), which corresponds to the number
// of forward pointers the node has. When a new element is inserted into the list,
// a node with a random level is inserted to represent the element. Random levels
// are generated with a simple pattern: 50% are level 1, 25% are level 2, 12.5% are
// level 3 and so on.
type Element[K Key] struct {
	key     K
	fingers []*Element[K]
}

// Value of element
func (el *Element[K]) Key() K { return el.key }

// Return next element in the set.
// Use for-loop to iterate through set elements
//
//	for e := set.Successor(...); e != nil; e.Next() { /* ... */}
func (el *Element[K]) Next() *Element[K] { return el.fingers[0] }

// Return next element in the set on level.
// Use for-loop to iterate through set elements
//
//	for e := set.ValuesOn(...); e != nil; e.NextOn(...) { /* ... */}
func (el *Element[K]) NextOn(level int) *Element[K] {
	if level >= len(el.fingers) {
		return nil
	}

	return el.fingers[level]
}

// Cast Element into string
func (el *Element[K]) String() string {
	fingers := ""
	for _, x := range el.fingers {
		if x != nil {
			fingers = fingers + fmt.Sprintf(" %v", x.key)
		} else {
			fingers = fingers + " _"
		}
	}

	return fmt.Sprintf("{ %4v\t|%s }", el.key, fingers)
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

	//
	ptable [L]float64

	// memory allocator for elements
	malloc Allocator[K, Element[K]]
}

// New create instance of SkipList
func NewSet[K Key](opts ...SetConfig[K]) *Set[K] {
	head := new(Element[K])
	head.fingers = make([]*Element[K], L)

	set := &Set[K]{
		head:   head,
		null:   *new(K),
		Length: 0,
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
		path:   [L]*Element[K]{},
		ptable: probabilityTable,
		malloc: nil,
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

// Max level of skip list
func (set *Set[K]) Level() int {
	for i := 0; i < L; i++ {
		if set.head.fingers[i] == nil {
			return i - 1
		}
	}
	return L - 1
}

// skip algorithm is similar to search algorithm that traversing forward pointers.
// skip maintain the vector path that contains a pointer to the rightmost node
// of level i or higher that is to the left of the location of the
// insertion/deletion.
func (set *Set[K]) skip(lvl int, key K) (*Element[K], [L]*Element[K]) {
	path := set.path

	node := set.head
	next := node.fingers
	for level := L - 1; level >= lvl; level-- {
		for next[level] != nil && next[level].key < key {
			node = node.fingers[level]
			next = node.fingers
		}
		path[level] = node
	}

	return next[lvl], path
}

// Add element to set, return true if element is new
func (set *Set[K]) Add(key K) bool {
	el, path := set.skip(0, key)

	if el != nil && el.key == key {
		return false
	}

	rank, el := set.createElement(L, key)

	// re-bind fingers to new node
	for level := 0; level < rank; level++ {
		el.fingers[level] = path[level].fingers[level]
		path[level].fingers[level] = el
	}

	set.Length++
	return true
}

// mkNode creates a new node, randomly defines empty fingers (level of the node)
func (set *Set[K]) createElement(maxL int, key K) (int, *Element[K]) {
	// See: https://golang.org/src/math/rand/rand.go#L150
	p := float64(set.random.Int63()) / (1 << 63)

	level := 0
	for level < maxL && p < set.ptable[level] {
		level++
	}

	var node *Element[K]
	if set.malloc == nil {
		node = &Element[K]{fingers: make([]*Element[K], level)}
	} else {
		node = set.malloc.Alloc(key)
	}
	node.key = key

	return level, node
}

// Check is element exists in set
func (set *Set[K]) Has(key K) bool {
	el, _ := set.skip(0, key)

	if el != nil && el.key == key {
		return true
	}

	return false
}

// Cut element from the set, returns true if element is removed
func (set *Set[K]) Cut(key K) bool {
	rank := L
	v, path := set.skip(0, key)

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
func (set *Set[K]) Successor(key K) *Element[K] {
	el, _ := set.skip(0, key)
	return el
}

// Split set of elements by key
func (set *Set[K]) Split(key K) *Set[K] {
	node, path := set.skip(0, key)

	for level, x := range path {
		x.fingers[level] = nil
	}

	head := new(Element[K])
	head.fingers = make([]*Element[K], L)

	tail := &Set[K]{
		head:   head,
		null:   *new(K),
		Length: 0,
		random: set.random,
		path:   [L]*Element[K]{},
		ptable: set.ptable,
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

// SetL[K] type projects Set[K] with all ops on level N
type SetL[K Key] Set[K]

func ToSetL[K Key](s *Set[K]) *SetL[K] { return (*SetL[K])(s) }

// Add element to set, return true if element is new
// The element would not be promoted higher than defined level
func (s *SetL[K]) Add(level int, key K) bool {
	set := (*Set[K])(s)
	el, path := set.skip(0, key)

	if el != nil && el.key == key {
		return false
	}

	rank, el := set.createElement(level, key)

	// re-bind fingers to new node
	for level := 0; level < rank; level++ {
		el.fingers[level] = path[level].fingers[level]
		path[level].fingers[level] = el
	}

	set.Length++
	return true
}

// Cut segment on the level
func (s *SetL[K]) Cut(level int, from *Element[K]) *Element[K] {
	set := (*Set[K])(s)

	to := from.NextOn(level)

	segment, path := set.skip(0, from.Next().key)

	var lastOnSegment *Element[K]

	if to != nil {
		_, pathToHi := set.skip(0, to.key)
		lastOnSegment = pathToHi[0]
	}

	for level := 0; level < L; level++ {
		if path[level] != nil {
			path[level].fingers[level] = to
		}
	}

	if to != nil {
		for i := 0; i < len(lastOnSegment.fingers); i++ {
			lastOnSegment.fingers[i] = nil
		}
	}

	return segment
}

// All set elements on defined level
func (s *SetL[K]) Values(level int) *Element[K] {
	if level >= L {
		return nil
	}

	set := (*Set[K])(s)
	return set.head.fingers[level]
}

// Successor elements from set on given level
func (s *SetL[K]) Successor(level int, key K) *Element[K] {
	set := (*Set[K])(s)
	el, _ := set.skip(level, key)
	return el
}

// Configure Set properties
type SetConfig[K Key] func(*Set[K])

// Configure Random Generator
func SetWithRandomSource[K Key](random rand.Source) SetConfig[K] {
	return func(set *Set[K]) {
		set.random = random
	}
}

// Configure Memory Allocator
func SetWithAllocator[K Key](malloc Allocator[K, Element[K]]) SetConfig[K] {
	return func(set *Set[K]) {
		set.malloc = malloc
	}
}

// Configure Probability table
// Use math.Log(B)/B < p < math.Pow(B, -0.5)
//
// The probability help to control the "distance" between elements on each level
// Use p = math.Pow(B, -0.5), where B is number of elements
// On L1 distance is √B, L2 distance is B, Ln distance is (√B)ⁿ
func SetWithProbability[K Key](p float64) SetConfig[K] {
	return func(set *Set[K]) {
		var ptable [L]float64

		for i := 1; i <= L; i++ {
			ptable[i-1] = math.Pow(p, float64(i-1))
		}

		set.ptable = ptable
	}
}

// Configure Probability table so that each level takes (√B)ⁿ elements
func SetWithBlockSize[K Key](b int) SetConfig[K] {
	return SetWithProbability[K](math.Pow(float64(b), -0.5))
}

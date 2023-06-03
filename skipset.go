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
	Key     K
	Fingers []*Element[K]
}

// Rank of node
func (el *Element[K]) Rank() int { return len(el.Fingers) }

// Return next element in the set.
// Use for-loop to iterate through set elements
//
//	for e := set.Successor(...); e != nil; e.Next() { /* ... */}
func (el *Element[K]) Next() *Element[K] { return el.Fingers[0] }

// Return next element in the set on level.
// Use for-loop to iterate through set elements
//
//	for e := set.ValuesOn(...); e != nil; e.NextOn(...) { /* ... */}
func (el *Element[K]) NextOn(level int) *Element[K] {
	if level >= len(el.Fingers) {
		return nil
	}

	return el.Fingers[level]
}

// Cast Element into string
func (el *Element[K]) String() string {
	fingers := ""
	for _, x := range el.Fingers {
		if x != nil {
			fingers = fingers + fmt.Sprintf(" %v", x.Key)
		} else {
			fingers = fingers + " _"
		}
	}

	return fmt.Sprintf("{ %4v\t|%s }", el.Key, fingers)
}

// --------------------------------------------------------------------------------------

// Set of Elements
type Set[K Key] struct {
	//
	// head of the list, the node is a lowest element
	head *Element[K]

	// null element of type T
	null K

	//
	// number of elements in the set, O(1)
	length int

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
	head := &Element[K]{Fingers: make([]*Element[K], L)}

	set := &Set[K]{
		head:   head,
		null:   *new(K),
		length: 0,
		random: rand.NewSource(time.Now().UnixNano()),
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
		v = v.Fingers[0]
	}

	return sb.String()
}

func (set *Set[K]) Length() int {
	return set.length
}

// Max level of skip list
func (set *Set[K]) Level() int {
	for i := 0; i < L; i++ {
		if set.head.Fingers[i] == nil {
			return i - 1
		}
	}
	return L - 1
}

// skip algorithm is similar to search algorithm that traversing forward pointers.
// skip maintain the vector path that contains a pointer to the rightmost node
// of level i or higher that is to the left of the location of the
// insertion/deletion.
func (set *Set[K]) Skip(level int, key K) (*Element[K], [L]*Element[K]) {
	path := set.path

	node := set.head
	next := node.Fingers
	for lev := L - 1; lev >= level; lev-- {
		for next[lev] != nil && next[lev].Key < key {
			node = node.Fingers[lev]
			next = node.Fingers
		}
		path[lev] = node
	}

	return next[level], path
}

// Add element to set, return true if element is new
func (set *Set[K]) Add(key K) (bool, *Element[K]) {
	el, path := set.Skip(0, key)

	if el != nil && el.Key == key {
		return false, el
	}

	rank, el := set.CreateElement(L, key)

	// re-bind fingers to new node
	for level := 0; level < rank; level++ {
		el.Fingers[level] = path[level].Fingers[level]
		path[level].Fingers[level] = el
	}

	set.length++
	return true, el
}

// mkNode creates a new node, randomly defines empty fingers (level of the node)
func (set *Set[K]) CreateElement(maxL int, key K) (int, *Element[K]) {
	// See: https://golang.org/src/math/rand/rand.go#L150
	p := float64(set.random.Int63()) / (1 << 63)

	level := 0
	for level < maxL && p < set.ptable[level] {
		level++
	}

	node := set.NewElement(key, level)
	node.Key = key

	return level, node
}

// allocate new node
func (set *Set[K]) NewElement(key K, rank int) *Element[K] {
	if set.malloc != nil {
		return set.malloc.Alloc(key)
	}

	return &Element[K]{Fingers: make([]*Element[K], rank)}
}

// Check is element exists in set
func (set *Set[K]) Has(key K) (bool, *Element[K]) {
	el, _ := set.Skip(0, key)

	if el != nil && el.Key == key {
		return true, el
	}

	return false, nil
}

// Cut element from the set, returns true if element is removed
func (set *Set[K]) Cut(key K) (bool, *Element[K]) {
	rank := L
	v, path := set.Skip(0, key)

	if v == nil || v.Key != key {
		return false, nil
	}

	for level := 0; level < rank; level++ {
		if path[level].Fingers[level] == v {
			if len(v.Fingers) > level {
				path[level].Fingers[level] = v.Fingers[level]
			} else {
				path[level].Fingers[level] = nil
			}
		}
	}

	set.length--

	if set.malloc != nil {
		set.malloc.Free(key)
	}

	return true, v
}

// Head of skiplist
func (set *Set[K]) Head() *Element[K] {
	return set.head
}

// All set elements
func (set *Set[K]) Values() *Element[K] {
	return set.head.Fingers[0]
}

// Successor elements of key
func (set *Set[K]) Successor(key K) *Element[K] {
	el, _ := set.Skip(0, key)
	return el
}

// Split set of elements by key
func (set *Set[K]) Split(key K) *Set[K] {
	node, path := set.Skip(0, key)

	for level, x := range path {
		x.Fingers[level] = nil
	}

	head := &Element[K]{Fingers: make([]*Element[K], L)}

	tail := &Set[K]{
		head:   head,
		null:   *new(K),
		length: 0,
		random: set.random,
		path:   [L]*Element[K]{},
		ptable: set.ptable,
		malloc: set.malloc,
	}
	tail.head.Fingers[0] = node

	length := 0
	for n := node; n != nil; n = n.Fingers[0] {
		length++
	}

	tail.length = length
	set.length -= length

	return tail
}

// --------------------------------------------------------------------------------------

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

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

// Each key-value pair is represented by a Pair in a skip structures. Each node has
// a height or level (length of fingers array), which corresponds to the number
// of forward pointers the node has. When a new element is inserted into the list,
// a node with a random level is inserted to represent the element. Random levels
// are generated with a simple pattern: 50% are level 1, 25% are level 2, 12.5% are
// level 3 and so on.
type Pair[K Key, V any] struct {
	Key     K
	Value   V
	Fingers []*Pair[K, V]
}

// Rank of node
func (el *Pair[K, V]) Rank() int { return len(el.Fingers) }

// Return next element in the set.
// Use for-loop to iterate through set elements
//
//	for e := set.Successor(...); e != nil; e.Next() { /* ... */}
func (el *Pair[K, V]) Next() *Pair[K, V] { return el.Fingers[0] }

// Return next element in the set on level.
// Use for-loop to iterate through set elements
//
//	for e := set.ValuesOn(...); e != nil; e.NextOn(...) { /* ... */}
func (el *Pair[K, V]) NextOn(level int) *Pair[K, V] {
	if level >= len(el.Fingers) {
		return nil
	}

	return el.Fingers[level]
}

// Cast Element into string
func (el *Pair[K, V]) String() string {
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

// Map of Elements
type Map[K Key, V any] struct {
	//
	// head of the list, the node is a lowest element
	head *Pair[K, V]

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
	path [L]*Pair[K, V]

	//
	ptable [L]float64

	// memory allocator for elements
	malloc Allocator[K, Pair[K, V]]
}

// New create instance of SkipList
func NewMap[K Key, V any](opts ...MapConfig[K, V]) *Map[K, V] {
	head := &Pair[K, V]{Fingers: make([]*Pair[K, V], L)}

	set := &Map[K, V]{
		head:   head,
		null:   *new(K),
		length: 0,
		random: rand.NewSource(time.Now().UnixNano()),
		path:   [L]*Pair[K, V]{},
		ptable: probabilityTable,
		malloc: nil,
	}

	for _, opt := range opts {
		opt(set)
	}

	return set
}

// Cast set into string
func (kv *Map[K, V]) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("--- SkipMap[%T] %p ---\n", kv.null, &kv))

	v := kv.head
	for v != nil {
		sb.WriteString(v.String())
		sb.WriteString("\n")
		v = v.Fingers[0]
	}

	return sb.String()
}

func (kv *Map[K, V]) Length() int {
	return kv.length
}

// Max level of skip list
func (kv *Map[K, V]) Level() int {
	for i := 0; i < L; i++ {
		if kv.head.Fingers[i] == nil {
			return i - 1
		}
	}
	return L - 1
}

// skip algorithm is similar to search algorithm that traversing forward pointers.
// skip maintain the vector path that contains a pointer to the rightmost node
// of level i or higher that is to the left of the location of the
// insertion/deletion.
func (kv *Map[K, V]) Skip(level int, key K) (*Pair[K, V], [L]*Pair[K, V]) {
	path := kv.path

	node := kv.head
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

func (kv *Map[K, V]) Put(key K, val V) (bool, *Pair[K, V]) {
	el, path := kv.Skip(0, key)

	if el != nil && el.Key == key {
		el.Value = val
		return false, el
	}

	rank, el := kv.CreatePair(L, key, val)

	// re-bind fingers to new node
	for level := 0; level < rank; level++ {
		el.Fingers[level] = path[level].Fingers[level]
		path[level].Fingers[level] = el
	}

	kv.length++
	return true, el
}

// creates a new node, randomly defines empty fingers (level of the node)
func (kv *Map[K, V]) CreatePair(maxL int, key K, val V) (int, *Pair[K, V]) {
	// See: https://golang.org/src/math/rand/rand.go#L150
	p := float64(kv.random.Int63()) / (1 << 63)

	level := 0
	for level < maxL && p < kv.ptable[level] {
		level++
	}

	node := kv.NewPair(key, level)
	node.Key = key
	node.Value = val

	return level, node
}

// allocate new pair
func (kv *Map[K, V]) NewPair(key K, rank int) *Pair[K, V] {
	if kv.malloc != nil {
		return kv.malloc.Alloc(key)
	}

	return &Pair[K, V]{Fingers: make([]*Pair[K, V], rank)}
}

// Check is element exists in set
func (kv *Map[K, V]) Get(key K) (V, *Pair[K, V]) {
	el, _ := kv.Skip(0, key)

	if el != nil && el.Key == key {
		return el.Value, el
	}

	return *new(V), nil
}

// Cut element from the set, returns true if element is removed
func (kv *Map[K, V]) Cut(key K) (bool, *Pair[K, V]) {
	rank := L
	v, path := kv.Skip(0, key)

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

	kv.length--

	if kv.malloc != nil {
		kv.malloc.Free(key)
	}

	return true, v
}

// Head of skiplist
func (kv *Map[K, V]) Head() *Pair[K, V] {
	return kv.head
}

// All set elements
func (kv *Map[K, V]) Values() *Pair[K, V] {
	return kv.head.Fingers[0]
}

// Successor elements from set
func (kv *Map[K, V]) Successor(key K) *Pair[K, V] {
	el, _ := kv.Skip(0, key)
	return el
}

// Split set of elements by key
func (kv *Map[K, V]) Split(key K) *Map[K, V] {
	node, path := kv.Skip(0, key)

	for level, x := range path {
		x.Fingers[level] = nil
	}

	head := &Pair[K, V]{Fingers: make([]*Pair[K, V], L)}

	tail := &Map[K, V]{
		head:   head,
		null:   *new(K),
		length: 0,
		random: kv.random,
		path:   [L]*Pair[K, V]{},
		ptable: kv.ptable,
		malloc: kv.malloc,
	}
	tail.head.Fingers[0] = node

	length := 0
	for n := node; n != nil; n = n.Fingers[0] {
		length++
	}

	tail.length = length
	kv.length -= length

	return tail
}

// --------------------------------------------------------------------------------------

// Configure Set properties
type MapConfig[K Key, V any] func(*Map[K, V])

// Configure Random Generator
func MapWithRandomSource[K Key, V any](random rand.Source) MapConfig[K, V] {
	return func(kv *Map[K, V]) {
		kv.random = random
	}
}

// Configure Memory Allocator
func MapWithAllocator[K Key, V any](malloc Allocator[K, Pair[K, V]]) MapConfig[K, V] {
	return func(kv *Map[K, V]) {
		kv.malloc = malloc
	}
}

// Configure Probability table
// Use math.Log(B)/B < p < math.Pow(B, -0.5)
//
// The probability help to control the "distance" between elements on each level
// Use p = math.Pow(B, -0.5), where B is number of elements
// On L1 distance is √B, L2 distance is B, Ln distance is (√B)ⁿ
func MapWithProbability[K Key, V any](p float64) MapConfig[K, V] {
	return func(kv *Map[K, V]) {
		var ptable [L]float64

		for i := 1; i <= L; i++ {
			ptable[i-1] = math.Pow(p, float64(i-1))
		}

		kv.ptable = ptable
	}
}

// Configure Probability table so that each level takes (√B)ⁿ elements
func MapWithBlockSize[K Key, V any](b int) MapConfig[K, V] {
	return MapWithProbability[K, V](math.Pow(float64(b), -0.5))
}

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
	key     K
	value   V
	fingers []*Pair[K, V]
}

// Value of element
func (el *Pair[K, V]) Key() K   { return el.key }
func (el *Pair[K, V]) Value() V { return el.value }

// Return next element in the set.
// Use for-loop to iterate through set elements
//
//	for e := set.Successor(...); e != nil; e.Next() { /* ... */}
func (el *Pair[K, V]) Next() *Pair[K, V] { return el.fingers[0] }

// Rank of node
func (el *Pair[K, V]) Rank() int { return len(el.fingers) }

// Return next element in the set on level.
// Use for-loop to iterate through set elements
//
//	for e := set.ValuesOn(...); e != nil; e.NextOn(...) { /* ... */}
func (el *Pair[K, V]) NextOn(level int) *Pair[K, V] {
	if level >= len(el.fingers) {
		return nil
	}

	return el.fingers[level]
}

// Cast Element into string
func (el *Pair[K, V]) String() string {
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
	head := &Pair[K, V]{fingers: make([]*Pair[K, V], L)}

	set := &Map[K, V]{
		head:   head,
		null:   *new(K),
		length: 0,
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
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
		v = v.fingers[0]
	}

	return sb.String()
}

func (kv *Map[K, V]) Length() int {
	return kv.length
}

// Max level of skip list
func (kv *Map[K, V]) Level() int {
	for i := 0; i < L; i++ {
		if kv.head.fingers[i] == nil {
			return i - 1
		}
	}
	return L - 1
}

// skip algorithm is similar to search algorithm that traversing forward pointers.
// skip maintain the vector path that contains a pointer to the rightmost node
// of level i or higher that is to the left of the location of the
// insertion/deletion.
func (kv *Map[K, V]) skip(lvl int, key K) (*Pair[K, V], [L]*Pair[K, V]) {
	path := kv.path

	node := kv.head
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

func (kv *Map[K, V]) Put(key K, val V) bool {
	el, path := kv.skip(0, key)

	if el != nil && el.key == key {
		el.value = val
		return false
	}

	rank, el := kv.createElement(L, key, val)

	// re-bind fingers to new node
	for level := 0; level < rank; level++ {
		el.fingers[level] = path[level].fingers[level]
		path[level].fingers[level] = el
	}

	kv.length++
	return true
}

// mkNode creates a new node, randomly defines empty fingers (level of the node)
func (kv *Map[K, V]) createElement(maxL int, key K, val V) (int, *Pair[K, V]) {
	// See: https://golang.org/src/math/rand/rand.go#L150
	p := float64(kv.random.Int63()) / (1 << 63)

	level := 0
	for level < maxL && p < kv.ptable[level] {
		level++
	}

	var node *Pair[K, V]
	if kv.malloc == nil {
		node = &Pair[K, V]{fingers: make([]*Pair[K, V], level)}
	} else {
		node = kv.malloc.Alloc(key)
	}
	node.key = key
	node.value = val

	return level, node
}

// Check is element exists in set
func (kv *Map[K, V]) Get(key K) (V, bool) {
	el, _ := kv.skip(0, key)

	if el != nil && el.key == key {
		return el.value, true
	}

	return *new(V), false
}

// Cut element from the set, returns true if element is removed
func (kv *Map[K, V]) Cut(key K) (V, bool) {
	rank := L
	v, path := kv.skip(0, key)

	if v == nil || v.key != key {
		return *new(V), false
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

	kv.length--

	if kv.malloc != nil {
		kv.malloc.Free(key)
	}

	return v.value, true
}

// All set elements
func (kv *Map[K, V]) Values() *Pair[K, V] {
	return kv.head.fingers[0]
}

// Successor elements from set
func (kv *Map[K, V]) Successor(key K) *Pair[K, V] {
	el, _ := kv.skip(0, key)
	return el
}

// Split set of elements by key
func (kv *Map[K, V]) Split(key K) *Map[K, V] {
	node, path := kv.skip(0, key)

	for level, x := range path {
		x.fingers[level] = nil
	}

	head := &Pair[K, V]{fingers: make([]*Pair[K, V], L)}

	tail := &Map[K, V]{
		head:   head,
		null:   *new(K),
		length: 0,
		random: kv.random,
		path:   [L]*Pair[K, V]{},
		ptable: kv.ptable,
		malloc: kv.malloc,
	}
	tail.head.fingers[0] = node

	length := 0
	for n := node; n != nil; n = n.fingers[0] {
		length++
	}

	tail.length = length
	kv.length -= length

	return tail
}

// --------------------------------------------------------------------------------------

// MapL[K, V] type projects Map[K, V] with all ops on level N
type MapL[K Key, V any] Map[K, V]

func ToMapL[K Key, V any](m *Map[K, V]) *MapL[K, V] { return (*MapL[K, V])(m) }

func (m *MapL[K, V]) PushH(seq []K) *Pair[K, V] {
	kv := (*Map[K, V])(m)

	for i := 1; i < len(seq); i++ {
		if kv.null != seq[i] {
			el, _ := kv.skip(0, seq[i])
			kv.head.fingers[i-1] = el
		}
	}

	return kv.head
}

// Explicitly create node with given topology
func (m *MapL[K, V]) Push(seq []K) *Pair[K, V] {
	kv := (*Map[K, V])(m)

	var node *Pair[K, V]
	if kv.malloc == nil {
		node = &Pair[K, V]{fingers: make([]*Pair[K, V], len(seq)-1)}
	} else {
		node = kv.malloc.Alloc(seq[0])
	}
	node.key = seq[0]

	for i := 1; i < len(seq); i++ {
		if kv.null != seq[i] {
			el, _ := kv.skip(0, seq[i])
			node.fingers[i-1] = el
		}
	}

	for i := 1; i < len(seq); i++ {
		kv.head.fingers[i-1] = node
	}

	return node
}

func (s *MapL[K, V]) Head() *Pair[K, V] {
	return s.head
}

// Add element to set, return true if element is new
// The element would not be promoted higher than defined level
func (m *MapL[K, V]) Put(level int, key K, val V) bool {
	kv := (*Map[K, V])(m)
	el, path := kv.skip(0, key)

	if el != nil && el.key == key {
		el.value = val
		return false
	}

	rank, el := kv.createElement(level, key, val)

	// re-bind fingers to new node
	for level := 0; level < rank; level++ {
		el.fingers[level] = path[level].fingers[level]
		path[level].fingers[level] = el
	}

	kv.length++
	return true
}

// Cut segment on the level
func (m *MapL[K, V]) Cut(level int, node *Pair[K, V]) *Pair[K, V] {
	if node == nil {
		return nil
	}

	kv := (*Map[K, V])(m)

	from := node
	if from == m.head.fingers[0] {
		// list.head is not available to client.
		// the cut of first segments should be started from head
		from = m.head
	}

	to := from.NextOn(level)
	segment := from.Next()

	// sometimes segment is equal to 0
	if segment == to {
		return nil
	}

	var lastOnSegment *Pair[K, V]

	if to != nil {
		_, pathToHi := kv.skip(0, to.key)
		lastOnSegment = pathToHi[0]
	}

	for i := 0; i < len(from.fingers); i++ {
		if from.fingers[i] != nil && (to == nil || from.fingers[i].key < to.key) {
			from.fingers[i] = to
		}
	}

	if to != nil {
		// detach last segment from list
		for i := 0; i < len(lastOnSegment.fingers); i++ {
			lastOnSegment.fingers[i] = nil
		}
	}

	length := 0
	for n := segment; n != nil; n = n.fingers[0] {
		length++
	}
	kv.length -= length

	return segment

}

// All set elements on defined level
func (m *MapL[K, V]) Values(level int) *Pair[K, V] {
	if level >= L {
		return nil
	}

	kv := (*Map[K, V])(m)
	return kv.head.fingers[level]
}

// Successor elements from set on given level
func (m *MapL[K, V]) Successor(level int, key K) *Pair[K, V] {
	kv := (*Map[K, V])(m)
	el, _ := kv.skip(level, key)
	return el
}

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

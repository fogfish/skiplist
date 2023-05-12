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
)

// Each element is represented by a tSkipNode in a skip list. Each node has
// a height or level (length of fingers array), which corresponds to the number
// of forward pointers the node has. When a new element is inserted into the list,
// a node with a random level is inserted to represent the element. Random levels
// are generated with a simple pattern: 50% are level 1, 25% are level 2, 12.5% are
// level 3 and so on.
type Node[K, V any] struct {
	key     K
	val     V
	fingers [L]*Node[K, V]
}

func newNode[K, V any](levels int) *Node[K, V] {
	fingers := [L]*Node[K, V]{}
	return &Node[K, V]{fingers: fingers}
}

func (node *Node[K, V]) String() string {
	fingers := ""
	for _, x := range node.fingers {
		if x != nil {
			fingers = fingers + fmt.Sprintf("%v ", x.key)
		} else {
			fingers = fingers + "nil "
		}
	}

	return fmt.Sprintf("{%v\t| %s}", node.key, fingers)
}

func (node *Node[K, V]) Key() K           { return node.key }
func (node *Node[K, V]) Value() V         { return node.val }
func (node *Node[K, V]) KeyValue() (K, V) { return node.key, node.val }

// Return next node, use for-loop to iterate through list
//
//	for node := skiplist.Lookup(...); node != nil; node.Next() { /* ... */}
func (node *Node[K, V]) Next() *Node[K, V] { return node.fingers[0] }

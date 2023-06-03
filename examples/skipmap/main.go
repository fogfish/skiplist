//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

package main

import (
	"fmt"

	"github.com/fogfish/skiplist"
)

func main() {
	skipmap := skiplist.NewMap[int, string]()

	// Put new values
	skipmap.Put(50, "instance")
	skipmap.Put(30, "a")
	skipmap.Put(10, "this")
	skipmap.Put(20, "is")
	skipmap.Put(60, "of")
	skipmap.Put(40, "new")
	skipmap.Put(70, "skipmap")

	// Debug skipset structure
	fmt.Println(skipmap)

	// Get values
	val, node := skipmap.Get(70)
	fmt.Printf("==> value by (70) exists: %v %v\n", val, node)

	val, node = skipmap.Get(35)
	fmt.Printf("==> value by (35) exists: %v %v\n", val, node)

	// Remove values
	fmt.Println("\n==> remove (40) value")
	skipmap.Cut(40)
	fmt.Println(skipmap)

	// values
	fmt.Println("\n==> values")
	for e := skipmap.Values(); e != nil; e = e.Next() {
		fmt.Printf(" (%d, %s)", e.Key, e.Value)
	}
	fmt.Println()

	// successors
	fmt.Println("\n==> successors (35)")
	for e := skipmap.Successor(35); e != nil; e = e.Next() {
		val, _ := skipmap.Get(e.Key)
		fmt.Printf(" (%d, %s)", e.Key, val)
	}
	fmt.Println()

	// split
	fmt.Println("\n==> split by (35)")
	tail := skipmap.Split(35)
	fmt.Println(skipmap)
	fmt.Println(tail)
}

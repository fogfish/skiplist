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
	skipset := skiplist.NewSet[string]()

	// Put new values
	skipset.Add("instance")
	skipset.Add("a")
	skipset.Add("this")
	skipset.Add("is")
	skipset.Add("of")
	skipset.Add("new")
	skipset.Add("skipset")

	// Debug skipset structure
	fmt.Println(skipset)

	// Get values
	fmt.Printf("==> value (skipset) exists: %v\n", skipset.Has("skipset"))
	fmt.Printf("==> value (rockset) exists: %v\n", skipset.Has("rockset"))

	// Remove values
	fmt.Println("\n==> remove (new) value")
	skipset.Cut("new")
	fmt.Println(skipset)

	// values
	fmt.Println("\n==> values")
	for e := skipset.Values(); e != nil; e = e.Next() {
		fmt.Printf(" %s", e.Key())
	}
	fmt.Println()

	// successors
	fmt.Println("\n==> successors (of)")
	for e := skipset.Successors("of"); e != nil; e = e.Next() {
		fmt.Printf(" %s", e.Key())
	}
	fmt.Println()

	// split
	fmt.Println("\n==> split by (of)")
	tail := skipset.Split("of")
	fmt.Println(skipset)
	fmt.Println(tail)
}

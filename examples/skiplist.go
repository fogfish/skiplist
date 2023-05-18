package main

import (
	"fmt"

	"github.com/fogfish/skiplist"
	"github.com/fogfish/skiplist/ord"
)

func main() {
	// each instance of SkipList requires "ordering" type law,
	// as defined by ord.Ord interface. The application might define own law or
	// use the default one, available for most of built-in types
	list := skiplist.New[int, string](ord.Type[int]())

	// Put new values
	fmt.Println("\n==> put values")
	skiplist.Put(list, 50, "instance")
	skiplist.Put(list, 30, "a")
	skiplist.Put(list, 10, "this")
	skiplist.Put(list, 20, "is")
	skiplist.Put(list, 60, "of")
	skiplist.Put(list, 40, "new")
	skiplist.Put(list, 70, "skiplist")

	// Debug skiplist structure
	fmt.Println(list)

	// Get values
	fmt.Printf("\n==> get values\n%s %s\n",
		skiplist.Get(list, 40),
		skiplist.Get(list, 70),
	)

	// Lookup
	fmt.Printf("\n==> lookup node\n%v\n",
		skiplist.Lookup(list, 40),
	)

	// Lookup before key
	fmt.Printf("\n==> lookup before\n%v\n",
		skiplist.LookupBefore(list, 35),
	)

	// Lookup after key
	fmt.Printf("\n==> lookup after\n%v\n",
		skiplist.LookupAfter(list, 55),
	)

	// Split the list by key
	fmt.Println("\n==> split list")
	a, b := skiplist.Split(list, 35)
	show(a)
	show(b)

	// Take While
	fmt.Println("\n==> take while < 55")
	c := skiplist.TakeWhile(skiplist.Values(list),
		func(k int, v string) bool { return k < 55 },
	)
	show(c)

	// Drop While
	fmt.Println("\n==> drop while < 35")
	d := skiplist.DropWhile(skiplist.Values(list),
		func(k int, v string) bool { return k < 35 },
	)
	show(d)

	// Range
	fmt.Println("\n==> take range [35, 60]")
	e := skiplist.Range(list, 35, 60)
	show(e)

	// Filter
	fmt.Println("\n==> filter")
	f := skiplist.Filter(skiplist.Values(list),
		func(k int, v string) bool { return len(v) > 2 },
	)
	show(f)

	// Remove values
	skiplist.Remove(list, 40)

	// Split list
	fmt.Println("\n==> split")
	head, tail := skiplist.SplitAt(list, 45)

	fmt.Println("head")
	fmt.Println(head)

	fmt.Println("tail")
	fmt.Println(tail)

}

func show(seq skiplist.Iterator[int, string]) {
	skiplist.FMap(seq, func(i int, s string) error {
		fmt.Printf("(%d %s) ", i, s)
		return nil
	})
	fmt.Println()
}

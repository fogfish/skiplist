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
	skiplist.Put(list, 5, "instance")
	skiplist.Put(list, 3, "a")
	skiplist.Put(list, 1, "this")
	skiplist.Put(list, 2, "is")
	skiplist.Put(list, 6, "of")
	skiplist.Put(list, 4, "new")
	skiplist.Put(list, 7, "skiplist")

	// Debug skiplist structure
	fmt.Println(list)

	// Get values
	fmt.Printf("%s %s\n",
		skiplist.Get(list, 4),
		skiplist.Get(list, 7),
	)

	// Remove values
	skiplist.Remove(list, 4)

	// Split the list by key
	a, b := skiplist.Split(list, 4)
	a.FMap(func(i int, s string) error {
		fmt.Printf("%s ", s)
		return nil
	})

	b.FMap(func(i int, s string) error {
		fmt.Printf("%s ", s)
		return nil
	})
	fmt.Println()
}

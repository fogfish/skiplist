//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

package skiplist_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/skiplist"
)

func ForSuite[K skiplist.Num, V any](
	t *testing.T,
	seq []K,
	gen func(K) skiplist.Iterator[K, V],
) {

	t.Run("For", func(t *testing.T) {
		for _, k := range []int{0, len(seq) / 4, len(seq) / 2, len(seq) - 1} {

			i := k
			e := gen(seq[k])
			for has := e != nil; has; has = e.Next() {
				it.Then(t).Should(
					it.Equal(e.Key(), seq[i]),
				)
				i++
			}
			it.Then(t).Should(it.Equal(i, len(seq)))
		}
	})

	t.Run("TakeWhile", func(t *testing.T) {
		for _, k := range [][]int{
			{0, 0},
			{0, len(seq) / 4},
			{len(seq) / 4, len(seq) / 2},
			{len(seq) / 2, len(seq) - 1},
			{len(seq) - 1, len(seq) - 1},
		} {
			i := k[0]
			e := skiplist.TakeWhile(gen(seq[i]),
				func(key K, val V) bool { return key < seq[k[1]] },
			)
			for has := e != nil; has; has = e.Next() {
				it.Then(t).Should(
					it.Equal(e.Key(), seq[i]),
				)
				i++
			}

			it.Then(t).Should(it.Equal(i, k[1]))
		}
	})

	t.Run("DropWhile", func(t *testing.T) {
		for _, k := range [][]int{
			{0, 0},
			{0, len(seq) / 4},
			{len(seq) / 4, len(seq) / 2},
			{len(seq) / 2, len(seq) - 1},
			{len(seq) - 1, len(seq) - 1},
		} {
			i := k[1]
			e := skiplist.DropWhile(gen(seq[i]),
				func(key K, val V) bool { return key < seq[k[1]] },
			)
			for has := e != nil; has; has = e.Next() {
				it.Then(t).Should(
					it.Equal(e.Key(), seq[i]),
				)
				i++
			}

			it.Then(t).Should(it.Equal(i, len(seq)))
		}
	})

	t.Run("Filter", func(t *testing.T) {
		for _, k := range []int{0, len(seq) / 4, len(seq) / 2, len(seq) - 1} {

			e := skiplist.Filter(gen(seq[k]),
				func(key K, val V) bool { return key%2 == 0 },
			)
			for has := e != nil; has; has = e.Next() {
				it.Then(t).Should(
					it.Equal(e.Key()%2, 0),
				)
			}
		}
	})

	t.Run("ForEach", func(t *testing.T) {
		for _, k := range []int{0, len(seq) / 4, len(seq) / 2, len(seq) - 1} {

			i := 0
			err := skiplist.ForEach(gen(seq[k]),
				func(key K, val V) error {
					i++
					return nil
				},
			)

			it.Then(t).Should(
				it.Equal(i, len(seq)-k),
				it.Nil(err),
			)
		}
	})

	t.Run("FMap", func(t *testing.T) {
		for _, k := range []int{0, len(seq) / 4, len(seq) / 2, len(seq) - 1} {

			i := k
			e := skiplist.FMap(gen(seq[k]),
				func(key K, val V) string { return fmt.Sprintf("%v|%v", key, val) },
			)
			for has := e != nil; has; has = e.Next() {
				it.Then(t).Should(
					it.Equal(e.Key(), seq[i]),
					it.Equal(e.Value(), fmt.Sprintf("%v|%v", seq[i], seq[i])),
				)
				i++
			}
			it.Then(t).Should(it.Equal(i, len(seq)))
		}
	})
}

func TestForSet(t *testing.T) {
	seq := []uint32{0x67, 0xaa, 0xb2, 0xd9, 0x56, 0xbd, 0x7c, 0xc6, 0x21, 0xaf, 0x22, 0xcf, 0xb1, 0x69, 0xcb, 0xa8}

	set := skiplist.NewSet[uint32]()
	for _, x := range seq {
		set.Add(x)
	}

	sort.Slice(seq, func(i, j int) bool { return seq[i] < seq[j] })

	ForSuite(t, seq,
		func(key uint32) skiplist.Iterator[uint32, uint32] {
			return skiplist.ForSet(set, set.Successors(key))
		},
	)
}

func TestForMap(t *testing.T) {
	seq := []uint32{0x67, 0xaa, 0xb2, 0xd9, 0x56, 0xbd, 0x7c, 0xc6, 0x21, 0xaf, 0x22, 0xcf, 0xb1, 0x69, 0xcb, 0xa8}

	kv := skiplist.NewMap[uint32, uint32]()
	for _, x := range seq {
		kv.Put(x, x)
	}

	sort.Slice(seq, func(i, j int) bool { return seq[i] < seq[j] })

	ForSuite(t, seq,
		func(key uint32) skiplist.Iterator[uint32, uint32] {
			return skiplist.ForMap(kv, kv.Successors(key))
		},
	)
}

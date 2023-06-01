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

	tseq "github.com/fogfish/golem/trait/seq"
	"github.com/fogfish/it/v2"
	"github.com/fogfish/skiplist"
)

func ForSuite[K skiplist.Num](
	t *testing.T,
	seq []K,
	gen func(K) tseq.Seq[K],
) {

	t.Run("For", func(t *testing.T) {
		for _, k := range []int{0, len(seq) / 4, len(seq) / 2, len(seq) - 1} {

			i := k
			e := gen(seq[k])
			for has := e != nil; has; has = e.Next() {
				it.Then(t).Should(
					it.Equal(e.Value(), seq[i]),
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
			e := tseq.TakeWhile(gen(seq[i]),
				func(key K) bool { return key < seq[k[1]] },
			)
			for has := e != nil; has; has = e.Next() {
				it.Then(t).Should(
					it.Equal(e.Value(), seq[i]),
				)
				i++
			}
			if e != nil {
				it.Then(t).ShouldNot(
					it.True(e.Next()),
					it.True(e.Next()),
				)
			}

			it.Then(t).Should(
				it.Equal(i, k[1]),
			)
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
			e := tseq.DropWhile(gen(seq[i]),
				func(key K) bool { return key < seq[k[1]] },
			)
			for has := e != nil; has; has = e.Next() {
				it.Then(t).Should(
					it.Equal(e.Value(), seq[i]),
				)
				i++
			}

			it.Then(t).Should(it.Equal(i, len(seq)))
		}
	})

	t.Run("Filter", func(t *testing.T) {
		for _, k := range []int{0, len(seq) / 4, len(seq) / 2, len(seq) - 1} {

			e := tseq.Filter(gen(seq[k]),
				func(key K) bool { return key%2 == 0 },
			)
			for has := e != nil; has; has = e.Next() {
				it.Then(t).Should(
					it.Equal(e.Value()%2, 0),
				)
			}
			if e != nil {
				it.Then(t).ShouldNot(
					it.True(e.Next()),
					it.True(e.Next()),
				)
			}
		}
	})

	t.Run("ForEach", func(t *testing.T) {
		for _, k := range []int{0, len(seq) / 4, len(seq) / 2, len(seq) - 1} {

			i := 0
			err := tseq.ForEach(gen(seq[k]),
				func(key K) error {
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
			e := tseq.Map(gen(seq[k]),
				func(key K) string { return fmt.Sprintf("%v", key) },
			)
			for has := e != nil; has; has = e.Next() {
				it.Then(t).Should(
					it.Equal(e.Value(), fmt.Sprintf("%v", seq[i])),
				)
				i++
			}
			it.Then(t).Should(it.Equal(i, len(seq)))
		}
	})

	t.Run("Plus", func(t *testing.T) {
		for _, k := range []int{0, len(seq) / 4, len(seq) / 2, len(seq) - 1} {

			m := len(seq) - k
			i := 0
			e := tseq.Plus(gen(seq[k]), gen(seq[k]))
			for has := e != nil; has; has = e.Next() {
				it.Then(t).Should(
					it.Equal(e.Value(), seq[k+i%m]),
				)
				i++
			}

			v := gen(seq[k])
			it.Then(t).Should(
				it.Equal(k+i/2, len(seq)),
				it.Equiv(tseq.Plus(v, nil), v),
				it.Equiv(tseq.Plus(nil, v), v),
			)
		}
	})

	t.Run("Join", func(t *testing.T) {
		for _, k := range []int{0, len(seq) / 4, len(seq) / 2, len(seq) - 1} {

			i := k
			e := tseq.Join(gen(seq[k]),
				func(k1 K) tseq.Seq[K] {
					return tseq.TakeWhile(gen(k1), func(k2 K) bool { return k1 == k2 })
				},
			)
			for has := e != nil; has; has = e.Next() {
				it.Then(t).Should(
					it.Equal(e.Value(), seq[i]),
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
		func(key uint32) tseq.Seq[uint32] {
			return skiplist.ForSet(set, set.Successor(key))
		},
	)

	t.Run("Nil", func(t *testing.T) {
		it.Then(t).Should(
			it.Nil(skiplist.ForSet(set, nil)),
		)
	})
}

func TestForMap(t *testing.T) {
	seq := []uint32{0x67, 0xaa, 0xb2, 0xd9, 0x56, 0xbd, 0x7c, 0xc6, 0x21, 0xaf, 0x22, 0xcf, 0xb1, 0x69, 0xcb, 0xa8}

	kv := skiplist.NewHashMap[uint32, uint32]()
	for _, x := range seq {
		kv.Put(x, x)
	}

	sort.Slice(seq, func(i, j int) bool { return seq[i] < seq[j] })

	ForSuite(t, seq,
		func(key uint32) tseq.Seq[uint32] {
			return skiplist.ForMap[uint32, uint32](kv, kv.Successors(key))
		},
	)

	t.Run("Nil", func(t *testing.T) {
		it.Then(t).Should(
			it.Nil(skiplist.ForMap[uint32, uint32](kv, nil)),
		)
	})
}

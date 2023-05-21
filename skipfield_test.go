//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

package skiplist_test

import (
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/skiplist"
)

func TestField(t *testing.T) {
	gf2 := skiplist.NewGF2[uint8]()
	key := uint8(0x39)

	for _, x := range [][]uint8{
		{0x00, 0x7f, 0xff},
		{0x00, 0x3f, 0x7f},
		{0x00, 0x1f, 0x3f},
		{0x20, 0x2f, 0x3f},
		{0x30, 0x37, 0x3f},
		{0x38, 0x3b, 0x3f},
		{0x38, 0x39, 0x3b},
		{0x38, 0x38, 0x39},
		{0x39, 0x39, 0x39},
	} {
		arc, _ := gf2.Get(key)
		it.Then(t).Should(
			it.Equal(arc.Lo, x[0]),
			it.Equal(arc.Hi, x[2]),
		)

		hd, tl := gf2.Add(key)
		it.Then(t).Should(
			it.Equal(hd.Lo, x[0]),
			it.Equal(hd.Hi, x[1]),
			it.Equal(tl.Hi, x[2]),
		)
	}

	topo := []uint8{0x1f, 0x2f, 0x37, 0x38, 0x39, 0x3b, 0x3f, 0x7f, 0xff}
	e := skiplist.ForGF2(gf2, gf2.Keys())

	for i := 0; i < len(topo); i++ {
		it.Then(t).Should(
			it.Equal(e.Key(), topo[i]),
		)
		e.Next()
	}

	e = skiplist.ForGF2(gf2, gf2.Successors(0x31))
	for i := 2; i < len(topo); i++ {
		it.Then(t).Should(
			it.Equal(e.Key(), topo[i]),
		)
		e.Next()
	}

	it.Then(t).Should(
		it.String(gf2.String()).Contain("SkipGF2"),
	)
}

func TestFieldPut(t *testing.T) {
	gf2 := skiplist.NewGF2[uint8]()
	gf2.Put(skiplist.Arc[uint8]{Rank: 7, Lo: 0, Hi: 0x7f})
	gf2.Put(skiplist.Arc[uint8]{Rank: 7, Lo: 0x80, Hi: 0xff})

	arc, _ := gf2.Get(0x60)
	it.Then(t).Should(
		it.Equal(arc.Lo, 0x00),
		it.Equal(arc.Hi, 0x7f),
	)

	arc, _ = gf2.Get(0xa0)
	it.Then(t).Should(
		it.Equal(arc.Lo, 0x80),
		it.Equal(arc.Hi, 0xff),
	)
}

// go test -fuzz=FuzzGF2
func FuzzGF2(f *testing.F) {
	field := skiplist.NewGF2[uint32]()
	f.Add(uint32(1024))

	f.Fuzz(func(t *testing.T, key uint32) {
		hd, tl := field.Add(key)
		if !(hd.Lo < hd.Hi && hd.Hi < tl.Lo && tl.Lo < tl.Hi) {
			t.Errorf("invalid split hd = %v, tl = %v", hd, tl)
		}
	})
}

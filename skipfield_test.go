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
		lo, hi := gf2.Get(key)
		it.Then(t).Should(
			it.Equal(lo, x[0]),
			it.Equal(hi, x[2]),
		)

		lo, mi, hi := gf2.Add(key)
		it.Then(t).Should(
			it.Equal(lo, x[0]),
			it.Equal(mi, x[1]),
			it.Equal(hi, x[2]),
		)
	}
}

// go test -fuzz=FuzzGF
func FuzzGF2(f *testing.F) {
	field := skiplist.NewGF2[uint32]()
	f.Add(uint32(1024))

	f.Fuzz(func(t *testing.T, key uint32) {
		lo, mi, hi := field.Add(key)
		if lo > mi || mi > hi || lo > hi {
			t.Errorf("invalid split (%d, %d, %d)", lo, mi, hi)
		}
	})
}

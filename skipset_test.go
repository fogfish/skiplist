//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

package skiplist_test

import (
	"math/rand"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/skiplist"
)

// ---------------------------------------------------------------

func SetSuite[K skiplist.Key](t *testing.T, seq []K) {
	//
	sorted := make([]K, len(seq))
	for i, x := range seq {
		sorted[i] = x
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	//
	set := skiplist.NewSet[K]()

	t.Run("Add", func(t *testing.T) {
		for _, el := range seq {
			it.Then(t).Should(
				it.True(set.Add(el)),
			).ShouldNot(
				it.True(set.Add(el)),
			)
		}

		it.Then(t).Should(it.Equal(set.Length, len(seq)))
	})

	t.Run("Has", func(t *testing.T) {
		for _, el := range seq {
			it.Then(t).Should(
				it.True(set.Has(el)),
			)
		}
	})

	t.Run("Values", func(t *testing.T) {
		values := set.Values()
		for i := 0; i < len(sorted); i++ {
			it.Then(t).Should(
				it.Equal(values.Key(), sorted[i]),
			)
			values = values.Next()
		}
	})

	t.Run("Successor", func(t *testing.T) {
		for _, k := range []int{0, len(sorted) / 4, len(sorted) / 2, len(sorted) - 1} {
			values := set.Successors(sorted[k])
			for i := k; i < len(sorted); i++ {
				it.Then(t).Should(
					it.Equal(values.Key(), sorted[i]),
				)
				values = values.Next()
			}
		}
	})

	t.Run("Cut", func(t *testing.T) {
		for _, el := range seq {
			it.Then(t).Should(
				it.True(set.Cut(el)),
			).ShouldNot(
				it.True(set.Cut(el)),
			)
		}

		it.Then(t).Should(it.Equal(set.Length, 0))
	})

	t.Run("Split", func(t *testing.T) {
		for _, k := range []int{0, len(sorted) / 4, len(sorted) / 2, len(sorted) - 1} {
			head := skiplist.NewSet[K]()
			for _, x := range seq {
				head.Add(x)
			}
			tail := head.Split(sorted[k])

			hval := head.Values()
			for i := 0; i < k; i++ {
				it.Then(t).Should(
					it.Equal(hval.Key(), sorted[i]),
				)
				hval = hval.Next()
			}

			tval := tail.Values()
			for i := k; i < len(sorted); i++ {
				it.Then(t).Should(
					it.Equal(tval.Key(), sorted[i]),
				)
				tval = tval.Next()
			}
		}
	})

}

func TestSetOfIntAddHasCut(t *testing.T) {
	SetSuite(t, []int{0x67})
	SetSuite(t, []int{0x67, 0xaa})
	SetSuite(t, []int{0x67, 0xaa, 0xb2, 0xd9, 0x56, 0xbd, 0x7c, 0xc6, 0x21, 0xaf, 0x22, 0xcf, 0xb1, 0x69, 0xcb, 0xa8})
}

func TestSetOfUIntAddHasCut(t *testing.T) {
	SetSuite(t, []uint{0x67})
	SetSuite(t, []uint{0x67, 0xaa})
	SetSuite(t, []uint{0x67, 0xaa, 0xb2, 0xd9, 0x56, 0xbd, 0x7c, 0xc6, 0x21, 0xaf, 0x22, 0xcf, 0xb1, 0x69, 0xcb, 0xa8})
}

func TestSetOfStringAddHasCut(t *testing.T) {
	SetSuite(t, []string{"67"})
	SetSuite(t, []string{"67", "aa"})
	SetSuite(t, []string{"67", "aa", "b2", "d9", "56", "bd", "7c", "c6", "21", "af", "22", "cf", "b1", "69", "cb", "a8"})
}

// ---------------------------------------------------------------

func SetBench[K skiplist.Key](b *testing.B, gen func(int) K) {
	size := 1000000
	defSet := skiplist.NewSet[K]()
	defKey := make([]K, size, size)

	for i := 0; i < size; i++ {
		key := gen(i)
		defKey[i] = key
		defSet.Add(key)
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(defKey),
		func(i, j int) { defKey[i], defKey[j] = defKey[j], defKey[i] })

	b.Run("AddToTail", func(b *testing.B) {
		set := skiplist.NewSet[K]()

		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			set.Add(gen(n))
		}
	})

	b.Run("AddToHead", func(b *testing.B) {
		set := skiplist.NewSet[K]()

		b.ReportAllocs()
		b.ResetTimer()
		for n := b.N; n > 0; n-- {
			set.Add(gen(n))
		}
	})

	b.Run("AddToRand", func(b *testing.B) {
		set := skiplist.NewSet[K]()

		b.ReportAllocs()
		b.ResetTimer()
		for n := b.N; n > 0; n-- {
			set.Add(gen(rand.Intn(n)))
		}
	})

	b.Run("Has", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			defSet.Has(defKey[n%size])
		}
	})

	b.Run("Successors", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			defSet.Successors(defKey[n%size])
		}
	})

	b.Run("Successors16", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			e := defSet.Successors(defKey[n%size])
			for i := 0; i < 16 && e != nil; i++ {
				e = e.Next()
			}
		}
	})

	b.Run("Successors64", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			e := defSet.Successors(defKey[n%size])
			for i := 0; i < 64 && e != nil; i++ {
				e = e.Next()
			}
		}
	})

	b.Run("Successors128", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			e := defSet.Successors(defKey[n%size])
			for i := 0; i < 64 && e != nil; i++ {
				e = e.Next()
			}
		}
	})
}

func BenchmarkSetOfInt(b *testing.B) {
	SetBench(b, func(i int) int { return i })
}

func BenchmarkSetOfUInt(b *testing.B) {
	SetBench(b, func(i int) uint { return uint(i) })
}

func BenchmarkSetOfString(b *testing.B) {
	SetBench(b, func(i int) string { return strconv.Itoa(i) })
}

// ---------------------------------------------------------------

// go test -fuzz=FuzzSetAddCut
func FuzzSetAddHas(f *testing.F) {
	set := skiplist.NewSet[string]()
	f.Add("abc")

	f.Fuzz(func(t *testing.T, el string) {
		set.Add(el)
		if !set.Has(el) {
			t.Errorf("element %s should be found", el)
		}
	})
}

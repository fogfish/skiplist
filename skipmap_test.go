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

func MapSuite[K skiplist.Key](t *testing.T, seq []K) {
	//
	sorted := make([]K, len(seq))
	copy(sorted, seq)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	//
	kv := skiplist.NewMap[K, K]()

	t.Run("Put", func(t *testing.T) {
		for _, el := range seq {
			it.Then(t).Should(
				it.True(kv.Put(el, el)),
			).ShouldNot(
				it.True(kv.Put(el, *new(K))),
				it.True(kv.Put(el, el)),
			)
		}

		it.Then(t).Should(it.Equal(kv.Length, len(seq)))
	})

	t.Run("Get", func(t *testing.T) {
		for _, el := range seq {
			val, has := kv.Get(el)
			it.Then(t).Should(
				it.True(has),
				it.Equal(val, el),
			)
		}
	})

	t.Run("Keys", func(t *testing.T) {
		values := kv.Values()
		for i := 0; i < len(sorted); i++ {
			val, has := kv.Get(values.Key())
			it.Then(t).Should(
				it.True(has),
				it.Equal(val, sorted[i]),
				it.Equal(values.Key(), sorted[i]),
				it.Equal(values.Value(), sorted[i]),
			)
			values = values.Next()
		}
	})

	t.Run("Successor", func(t *testing.T) {
		for _, k := range []int{0, len(sorted) / 4, len(sorted) / 2, len(sorted) - 1} {
			values := kv.Successor(sorted[k])
			for i := k; i < len(sorted); i++ {
				val, has := kv.Get(values.Key())
				it.Then(t).Should(
					it.True(has),
					it.Equal(val, sorted[i]),
					it.Equal(values.Key(), sorted[i]),
				)
				values = values.Next()
			}
		}
	})

	t.Run("String", func(t *testing.T) {
		it.Then(t).Should(
			it.String(kv.String()).Contain("SkipMap"),
		)
	})

	t.Run("Cut", func(t *testing.T) {
		for _, el := range seq {
			val, has := kv.Cut(el)
			_, exist := kv.Cut(el)
			it.Then(t).Should(
				it.True(has),
				it.Equal(val, el),
			).ShouldNot(
				it.True(exist),
			)
		}

		it.Then(t).Should(it.Equal(kv.Length, 0))
	})

	t.Run("Split", func(t *testing.T) {
		for _, k := range []int{0, len(sorted) / 4, len(sorted) / 2, len(sorted) - 1} {
			head := skiplist.NewMap[K, K]()
			for _, x := range seq {
				head.Put(x, x)
			}
			tail := head.Split(sorted[k])

			hval := head.Values()
			for i := 0; i < k; i++ {
				val, has := head.Get(hval.Key())
				_, exist := tail.Get(hval.Key())
				it.Then(t).Should(
					it.True(has),
					it.Equal(val, sorted[i]),
					it.Equal(hval.Key(), sorted[i]),
				).ShouldNot(
					it.True(exist),
				)
				hval = hval.Next()
			}

			tval := tail.Values()
			for i := k; i < len(sorted); i++ {
				val, has := tail.Get(tval.Key())
				_, exist := head.Get(tval.Key())
				it.Then(t).Should(
					it.True(has),
					it.Equal(val, sorted[i]),
					it.Equal(tval.Key(), sorted[i]),
				).ShouldNot(
					it.True(exist),
				)
				tval = tval.Next()
			}
		}
	})

}

func TestMapOfIntPutGetCut(t *testing.T) {
	MapSuite(t, []int{0x67})
	MapSuite(t, []int{0x67, 0xaa})
	MapSuite(t, []int{0x67, 0xaa, 0xb2, 0xd9, 0x56, 0xbd, 0x7c, 0xc6, 0x21, 0xaf, 0x22, 0xcf, 0xb1, 0x69, 0xcb, 0xa8})
}

func TestMapOfUIntPutGetCut(t *testing.T) {
	MapSuite(t, []uint{0x67})
	MapSuite(t, []uint{0x67, 0xaa})
	MapSuite(t, []uint{0x67, 0xaa, 0xb2, 0xd9, 0x56, 0xbd, 0x7c, 0xc6, 0x21, 0xaf, 0x22, 0xcf, 0xb1, 0x69, 0xcb, 0xa8})
}

func TestMapOfStringPutGetCut(t *testing.T) {
	MapSuite(t, []string{"67"})
	MapSuite(t, []string{"67", "aa"})
	MapSuite(t, []string{"67", "aa", "b2", "d9", "56", "bd", "7c", "c6", "21", "af", "22", "cf", "b1", "69", "cb", "a8"})
}

// ---------------------------------------------------------------

func MapBench[K skiplist.Key](b *testing.B, gen func(int) K) {
	size := 1000000
	defMap := skiplist.NewMap[K, K]()
	defKey := make([]K, size)

	for i := 0; i < size; i++ {
		key := gen(i)
		defKey[i] = key
		defMap.Put(key, key)
	}

	rand.New(rand.NewSource(time.Now().UnixNano())).Shuffle(
		len(defKey),
		func(i, j int) { defKey[i], defKey[j] = defKey[j], defKey[i] },
	)

	b.Run("PutToTail", func(b *testing.B) {
		kv := skiplist.NewMap[K, K]()

		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			key := gen(n)
			kv.Put(key, key)
		}
	})

	b.Run("PutToHead", func(b *testing.B) {
		kv := skiplist.NewMap[K, K]()

		b.ReportAllocs()
		b.ResetTimer()
		for n := b.N; n > 0; n-- {
			key := gen(n)
			kv.Put(key, key)
		}
	})

	b.Run("PutToRand", func(b *testing.B) {
		kv := skiplist.NewMap[K, K]()

		b.ReportAllocs()
		b.ResetTimer()
		for n := b.N; n > 0; n-- {
			key := gen(rand.Intn(n))
			kv.Put(key, key)
		}
	})

	b.Run("Get", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			defMap.Get(defKey[n%size])
		}
	})

	b.Run("Successors", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			defMap.Successor(defKey[n%size])
		}
	})

	b.Run("Successors16", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			e := defMap.Successor(defKey[n%size])
			for i := 0; i < 16 && e != nil; i++ {
				defMap.Get(e.Key())
				e = e.Next()
			}
		}
	})

	b.Run("Successors64", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			e := defMap.Successor(defKey[n%size])
			for i := 0; i < 64 && e != nil; i++ {
				defMap.Get(e.Key())
				e = e.Next()
			}
		}
	})

	b.Run("Successors128", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			e := defMap.Successor(defKey[n%size])
			for i := 0; i < 64 && e != nil; i++ {
				defMap.Get(e.Key())
				e = e.Next()
			}
		}
	})
}

func BenchmarkMapOfInt(b *testing.B) {
	MapBench(b, func(i int) int { return i })
}

func BenchmarkMapOfUInt(b *testing.B) {
	MapBench(b, func(i int) uint { return uint(i) })
}

func BenchmarkMapOfString(b *testing.B) {
	MapBench(b, func(i int) string { return strconv.Itoa(i) })
}

// ---------------------------------------------------------------

// go test -fuzz=FuzzMapIntPutGet
func FuzzMapIntPutGet(f *testing.F) {
	kv := skiplist.NewMap[uint64, string]()
	f.Add(uint64(123), "abc")

	f.Fuzz(func(t *testing.T, key uint64, val string) {
		kv.Put(key, val)

		el := kv.Successor(key)
		if el == nil {
			t.Errorf("pair (%v, %v) should be found", key, val)
		}

		x, has := kv.Get(el.Key())
		if !has {
			t.Errorf("pair (%v, %v) should be found", key, val)
		}

		if x != val {
			t.Errorf("pair (%v, %v) should contain %v", key, x, val)
		}
	})
}

// go test -fuzz=FuzzMapStringPutGet
func FuzzMapStringPutGet(f *testing.F) {
	kv := skiplist.NewMap[string, uint64]()
	f.Add("abc", uint64(123))

	f.Fuzz(func(t *testing.T, key string, val uint64) {
		kv.Put(key, val)

		el := kv.Successor(key)
		if el == nil {
			t.Errorf("pair (%v, %v) should be found", key, val)
		}

		x, has := kv.Get(el.Key())
		if !has {
			t.Errorf("pair (%v, %v) should be found", key, val)
		}

		if x != val {
			t.Errorf("pair (%v, %v) should contain %v", key, x, val)
		}
	})

}

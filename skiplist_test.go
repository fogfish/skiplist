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
	"math/rand"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/skiplist"
	"github.com/fogfish/skiplist/ord"
)

func Suite[K comparable, V any](t *testing.T, ord ord.Ord[K], seed map[K]V) {
	keys := make([]K, 0, len(seed))
	for k := range seed {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return ord.Compare(keys[i], keys[j]) == -1 })

	nul := skiplist.New[K, V](ord, rand.NewSource(0))

	one := skiplist.New[K, V](ord)
	skiplist.Put(one, keys[0], seed[keys[0]])

	few := skiplist.New[K, V](ord)
	for k, v := range seed {
		skiplist.Put(few, k, v)
	}

	t.Run("String", func(t *testing.T) {
		it.Then(t).ShouldNot(
			it.Equal(len(one.String()), 0),
		)
	})

	t.Run("Length", func(t *testing.T) {
		it.Then(t).Should(
			it.Equal(skiplist.Length(nul), 0),
			it.Equal(skiplist.Length(one), 1),
			it.Equal(skiplist.Length(few), len(seed)),
		)
	})

	t.Run("Put", func(t *testing.T) {
		list := skiplist.New[K, V](ord)
		for k, v := range seed {
			skiplist.Put(list, k, v)

			it.Then(t).
				Should(it.Equiv(v, skiplist.Get(list, k)))
		}

		for k, v := range seed {
			it.Then(t).
				Should(it.Equiv(v, skiplist.Get(list, k)))
		}
	})

	t.Run("Get", func(t *testing.T) {
		key := keys[0]
		val := skiplist.Get(nul, key)
		it.Then(t).
			Should(it.Equiv(val, *new(V)))

		val = skiplist.Get(one, key)
		it.Then(t).
			Should(it.Equiv(val, seed[key]))

		val = skiplist.Get(few, key)
		it.Then(t).
			Should(it.Equiv(val, seed[key]))

		for k, v := range seed {
			it.Then(t).
				Should(it.Equiv(v, skiplist.Get(few, k)))
		}
	})

	t.Run("Lookup", func(t *testing.T) {
		for _, at := range []int{
			0,
			len(keys) / 4,
			len(keys) / 2,
			len(keys) - 1,
		} {
			key := keys[at]
			node := skiplist.Lookup(few, key)
			before := skiplist.LookupBefore(few, key)
			after := skiplist.LookupAfter(few, key)

			it.Then(t).Should(
				it.Equal(node.Key(), key),
				it.Equiv(node.Value(), seed[key]),
				it.Equal(before.Key(), key),
				it.Equiv(before.Value(), seed[key]),
				it.Equal(after.Key(), key),
				it.Equiv(after.Value(), seed[key]),
			)
		}
	})

	t.Run("Values", func(t *testing.T) {
		values := toSeq(few.Head().Seq())

		it.Then(t).Should(
			it.Seq(values).Equal(keys...),
		)
	})

	t.Run("Node.Next", func(t *testing.T) {
		seq := make([]K, 0)
		for node := few.Head(); node != nil; node = node.Next() {
			seq = append(seq, node.Key())

			it.Then(t).Should(
				it.Equiv(node.Value(), seed[node.Key()]),
			)
		}

		it.Then(t).Should(
			it.Seq(seq).Equal(keys...),
		)
	})

	t.Run("ValuesFMap", func(t *testing.T) {
		values := skiplist.Values(few)

		i := 0
		skiplist.FMap(values, func(k K, v V) error {
			i++
			it.Then(t).Should(
				it.Equiv(k, keys[i]),
				it.Equiv(v, seed[k]),
			)
			return nil
		})
	})

	t.Run("Split", func(t *testing.T) {
		for _, at := range []int{
			0,
			len(keys) / 4,
			len(keys) / 2,
			len(keys) - 1,
		} {
			key := keys[at]
			before, after := skiplist.Split(few, key)
			seqB := toSeq(before)
			seqA := toSeq(after)

			it.Then(t).Should(
				it.Seq(seqB).Equal(keys[:at]...),
				it.Seq(seqA).Equal(keys[at:]...),
			)
		}
	})

	t.Run("Range", func(t *testing.T) {
		for _, at := range [][]int{
			{0, len(keys) / 4},
			{len(keys) / 4, len(keys) / 2},
			{len(keys) / 2, len(keys) - 1},
		} {
			from, to := keys[at[0]], keys[at[1]]
			seq := toSeq(
				skiplist.Range(few, from, to),
			)

			it.Then(t).Should(
				it.Seq(seq).Equal(keys[at[0] : at[1]+1]...),
			)
		}
	})

	t.Run("TakeWhile", func(t *testing.T) {
		for _, at := range []int{
			0,
			1,
			len(keys) / 4,
			len(keys) / 2,
			len(keys) - 1,
		} {
			seq := toSeq(
				skiplist.TakeWhile(skiplist.Values(few),
					func(k K, v V) bool { return ord.Compare(k, keys[at]) == -1 },
				),
			)

			it.Then(t).Should(
				it.Seq(seq).Equal(keys[:at]...),
			)
		}
	})

	t.Run("DropWhile", func(t *testing.T) {
		for _, at := range []int{
			0,
			1,
			len(keys) / 4,
			len(keys) / 2,
			len(keys) - 1,
		} {
			seq := toSeq(
				skiplist.DropWhile(skiplist.Values(few),
					func(k K, v V) bool { return ord.Compare(k, keys[at]) == -1 },
				),
			)

			it.Then(t).Should(
				it.Seq(seq).Equal(keys[at:]...),
			)
		}
	})

	t.Run("Filter", func(t *testing.T) {
		for _, at := range [][]int{
			{0, len(keys) / 4},
			{len(keys) / 4, len(keys) / 2},
			{len(keys) / 2, len(keys) - 1},
		} {
			from, to := keys[at[0]], keys[at[1]]
			seq := toSeq(
				skiplist.Filter(skiplist.Values(few),
					func(k K, v V) bool { return ord.Compare(k, from) > -1 && ord.Compare(k, to) < 1 },
				),
			)

			it.Then(t).Should(
				it.Seq(seq).Equal(keys[at[0] : at[1]+1]...),
			)
		}
	})

	t.Run("Map", func(t *testing.T) {
		seq := make([]K, 0)
		itr := skiplist.Map(few.Head().Seq(), func(k K, v V) K { return k })

		for has := itr != nil; has; has = itr.Next() {
			_, v := itr.KeyValue()
			seq = append(seq, v)
		}

		it.Then(t).Should(
			it.Seq(seq).Equal(keys...),
		)
	})

	t.Run("Remove", func(t *testing.T) {
		key := keys[0]
		val0 := skiplist.Remove(nul, key)
		val1 := skiplist.Get(nul, key)
		it.Then(t).Should(
			it.Equiv(val0, *new(V)),
			it.Equiv(val1, *new(V)),
		)

		val0 = skiplist.Remove(one, key)
		val1 = skiplist.Get(one, key)
		it.Then(t).Should(
			it.Equiv(val0, seed[key]),
			it.Equiv(val1, *new(V)),
		)

		val0 = skiplist.Remove(few, key)
		val1 = skiplist.Get(few, key)
		it.Then(t).Should(
			it.Equiv(val0, seed[key]),
			it.Equiv(val1, *new(V)),
		)
	})
}

func toSeq[K, V any](seq skiplist.Iterator[K, V]) []K {
	keys := make([]K, 0)

	for has := seq != nil; has; has = seq.Next() {
		keys = append(keys, seq.Key())
	}

	return keys
}

func Bench[K, V comparable](b *testing.B, compare ord.Ord[K], gen func(int) (K, V)) {
	var (
		rnd                                    = rand.New(rand.NewSource(time.Now().UnixNano()))
		defCap        int                      = 1000000
		defMapLike    map[K]V                  = make(map[K]V)
		defSkipList   *skiplist.SkipList[K, V] = skiplist.New[K, V](compare)
		defShuffleKey []K                      = make([]K, defCap)
		defShuffleVal []V                      = make([]V, defCap)
	)

	for i := 0; i < defCap; i++ {
		key, val := gen(i)

		skiplist.Put(defSkipList, key, val)
		defMapLike[key] = val

		rndKey, rndVal := gen(rnd.Intn(defCap))
		defShuffleKey[i] = rndKey
		defShuffleVal[i] = rndVal
	}

	b.Run("PutTail", func(b *testing.B) {
		list := skiplist.New[K, V](compare)

		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			key, val := gen(n)
			skiplist.Put(list, key, val)
		}
	})

	b.Run("PutHead", func(b *testing.B) {
		list := skiplist.New[K, V](compare)

		b.ReportAllocs()
		b.ResetTimer()
		for n := b.N; n > 0; n-- {
			key, val := gen(n)
			skiplist.Put(list, key, val)
		}
	})

	b.Run("PutRand", func(b *testing.B) {
		list := skiplist.New[K, V](compare)

		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			key := defShuffleKey[n%defCap]
			val := defShuffleVal[n%defCap]
			skiplist.Put(list, key, val)
		}
	})

	b.Run("GetRand", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			key := defShuffleKey[n%defCap]
			val := defShuffleVal[n%defCap]
			vxx := skiplist.Get(defSkipList, key)
			if val != vxx {
				panic(fmt.Errorf("invalid state for key %v, unexpected %v", key, val))
			}
		}
	})

	b.Run("GetRandMapLike", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			key := defShuffleKey[n%defCap]
			val := defShuffleVal[n%defCap]
			vxx := defMapLike[key]
			if val != vxx {
				panic(fmt.Errorf("invalid state for key %v, unexpected %v", key, val))
			}
		}
	})

}

func TestSkipListIntString(t *testing.T) {
	seed := map[int]string{}
	for i := 1; i < 1000; i++ {
		seed[i] = strconv.Itoa(i)
	}

	Suite[int](t, ord.Int, seed)
}

func TestSkipListStringStringPtr(t *testing.T) {
	seed := map[string]*string{}
	for i := 1; i < 1000; i++ {
		seed[strconv.Itoa(i)] = ptrOf(strconv.Itoa(i))
	}

	Suite[string](t, ord.String, seed)
}

func TestSkipListStringPtrStringPtr(t *testing.T) {
	seed := map[*string]*string{}
	for i := 1; i < 1000; i++ {
		seed[ptrOf(strconv.Itoa(i))] = ptrOf(strconv.Itoa(i))
	}

	cmp := ord.From[*string](
		func(a, b *string) int { return ord.String.Compare(*a, *b) },
	)

	Suite[*string](t, cmp, seed)
}

func ptrOf[T any](v T) *T { return &v }

func BenchmarkSkipListIntString(b *testing.B) {
	Bench[int](b,
		ord.Int,
		func(i int) (int, int) { return i, i },
	)
}

// func BenchmarkSkipListStringString(b *testing.B) {
// 	Bench[string](b,
// 		ord.String,
// 		func(i int) (string, string) {
// 			s := strconv.Itoa(i)
// 			return s, s
// 		},
// 	)
// }

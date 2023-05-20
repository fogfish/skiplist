//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

// Package skiplist implements a probabilistic list-based data structure
// that are a simple and efficient substitute for balanced trees.
//
// Please see the article that depicts the data structure
// https://15721.courses.cs.cmu.edu/spring2018/papers/08-oltpindexes1/pugh-skiplists-cacm1990.pdf
// http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.17.524
package skiplist

// L depth of fingers at each node.
//
// The value is estimated as math.Log10(float64(n)) / math.Log10(1/p)
// n = 4294967296, p = 1/math.E
const L = 22

// The probability table is generated for L=22
var probabilityTable []float64 = []float64{1, 0.36787944117144233, 0.1353352832366127, 0.04978706836786395, 0.018315638888734182, 0.006737946999085468, 0.002478752176666359, 0.0009118819655545165, 0.0003354626279025119, 0.0001234098040866796, 4.539992976248486e-05, 1.6701700790245666e-05, 6.1442123533282115e-06, 2.260329406981055e-06, 8.315287191035682e-07, 3.0590232050182594e-07, 1.1253517471925916e-07, 4.139937718785168e-08, 1.5229979744712636e-08, 5.60279643753727e-09, 2.0611536224385587e-09, 7.582560427911911e-10, 0}

// Constraint on key types supported by the data structures
type Key interface {
	~string |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Memory allocator
type Allocator[K Key, T any] interface {
	Alloc(K) *T
	Free(K)
}

type malloc[K Key, T any] struct{}

func (malloc[K, T]) Alloc(K) *T { return new(T) }
func (malloc[K, T]) Free(K)     {}

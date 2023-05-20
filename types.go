//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

package skiplist

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

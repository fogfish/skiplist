//
// Copyright (C) 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/skiplist
//

package ord

/*

Ord : T ⟼ T ⟼ Ordering
Each type implements compare rules, mapping pair of value to enum{ -1, 0, 1 }
*/
type Ord[T any] interface{ Compare(T, T) int }

type Comparable interface {
	~string |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

/*

ord generic implementation for built-in types
*/
type ord[T Comparable] string

func (ord[T]) Compare(a, b T) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func Type[T Comparable]() Ord[T] { return ord[T]("") }

/*

From is a combinator that lifts T ⟼ T ⟼ Ordering function to
an instance of Ord type trait
*/
type From[T any] func(T, T) int

func (f From[T]) Compare(a, b T) int { return f(a, b) }

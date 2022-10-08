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

/*

ord for bytes
*/
type bytes string

func (bytes) Compare(a, b []byte) int {
	switch {
	case len(a) < len(b):
		return -1
	case len(a) > len(b):
		return 1
	default:
		for i, v := range a {
			switch {
			case v < b[i]:
				return -1
			case v > b[i]:
				return 1
			}
		}
		return 0
	}
}

/*

built-in ord.Ord
*/
const (
	String  = ord[string]("skiplist.ord.string")
	Int     = ord[int]("skiplist.ord.int")
	Int8    = ord[int8]("skiplist.ord.int8")
	Int16   = ord[int16]("skiplist.ord.int16")
	Int32   = ord[int32]("skiplist.ord.int32")
	Int64   = ord[int64]("skiplist.ord.int64")
	UInt    = ord[uint]("skiplist.ord.uint")
	UInt8   = ord[uint8]("skiplist.ord.uint8")
	UInt16  = ord[uint16]("skiplist.ord.uint16")
	UInt32  = ord[uint32]("skiplist.ord.uint32")
	UInt64  = ord[uint64]("skiplist.ord.uint64")
	Float32 = ord[float32]("skiplist.ord.float32")
	Float64 = ord[float64]("skiplist.ord.float64")
	Bytes   = bytes("bytes")
)

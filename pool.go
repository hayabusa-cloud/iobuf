// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package iobuf

// Pool is a generic object pool interface with configurable blocking semantics.
//
// Implementations may operate in blocking or non-blocking mode. In blocking
// mode, Get blocks until an item is available and Put blocks until space
// is available. In non-blocking mode, both operations return iox.ErrWouldBlock
// instead of blocking.
//
// All implementations must be safe for concurrent use.
type Pool[T any] interface {
	// Put returns the item to the pool.
	// Returns iox.ErrWouldBlock if non-blocking and full.
	Put(item T) error

	// Get acquires an item from the pool.
	// Returns iox.ErrWouldBlock if non-blocking and empty.
	Get() (item T, err error)
}

// IndirectPool manages fixed-size buffer arrays by index. A successful Get
// gives the caller one slot. Return its index once with Put after all use of
// the slot has ended.
type IndirectPool[T BufferType] interface {
	Pool[int]

	// Value returns a copy of the array at the caller's acquired index.
	Value(indirect int) T

	// SetValue stores the array at the caller's acquired index.
	SetValue(indirect int, item T)
}

type (
	// PicoBufferPool manages 32-byte buffers via indirect indexing.
	PicoBufferPool = IndirectPool[PicoBuffer]

	// NanoBufferPool manages 128-byte buffers via indirect indexing.
	NanoBufferPool = IndirectPool[NanoBuffer]

	// MicroBufferPool manages 512-byte buffers via indirect indexing.
	MicroBufferPool = IndirectPool[MicroBuffer]

	// SmallBufferPool manages 2 KiB buffers via indirect indexing.
	SmallBufferPool = IndirectPool[SmallBuffer]

	// MediumBufferPool manages 8 KiB buffers via indirect indexing.
	MediumBufferPool = IndirectPool[MediumBuffer]

	// BigBufferPool manages 32 KiB buffers via indirect indexing.
	BigBufferPool = IndirectPool[BigBuffer]

	// LargeBufferPool manages 128 KiB buffers via indirect indexing.
	LargeBufferPool = IndirectPool[LargeBuffer]

	// GreatBufferPool manages 512 KiB buffers via indirect indexing.
	GreatBufferPool = IndirectPool[GreatBuffer]

	// HugeBufferPool manages 2 MiB buffers via indirect indexing.
	HugeBufferPool = IndirectPool[HugeBuffer]

	// VastBufferPool manages 8 MiB buffers via indirect indexing.
	VastBufferPool = IndirectPool[VastBuffer]

	// GiantBufferPool manages 32 MiB buffers via indirect indexing.
	GiantBufferPool = IndirectPool[GiantBuffer]

	// TitanBufferPool manages 128 MiB buffers via indirect indexing.
	TitanBufferPool = IndirectPool[TitanBuffer]
)

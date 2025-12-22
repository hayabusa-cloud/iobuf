// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package iobuf

// Pool is a generic object pool with blocking or non-blocking semantics.
type Pool[T any] interface {
	// Put returns the item to the pool. Returns iox.ErrWouldBlock
	// if non-blocking and full; blocks otherwise until space is available.
	Put(item T) error
	// Get acquires an item from the pool. Returns iox.ErrWouldBlock
	// if non-blocking and empty; blocks otherwise until an item is available.
	Get() (item T, err error)
}

// IndirectPool manages items by index rather than by value, enabling
// zero-copy access to pooled buffers via Value/SetValue.
type IndirectPool[T BufferType] interface {
	Pool[int]
	// Value returns the buffer associated with the given indirect index.
	Value(indirect int) T
	// SetValue updates the buffer at the specified indirect index.
	SetValue(indirect int, item T)
}

type (
	// PicoBufferPool manages 16-byte buffers via indirect indexing.
	PicoBufferPool = IndirectPool[PicoBuffer]

	// NanoBufferPool manages 64-byte buffers via indirect indexing.
	NanoBufferPool = IndirectPool[NanoBuffer]

	// MicroBufferPool manages 256-byte buffers via indirect indexing.
	MicroBufferPool = IndirectPool[MicroBuffer]

	// SmallBufferPool manages 1,024-byte (1 KiB) buffers via indirect indexing.
	SmallBufferPool = IndirectPool[SmallBuffer]

	// MediumBufferPool manages 4,096-byte (4 KiB) buffers via indirect indexing.
	MediumBufferPool = IndirectPool[MediumBuffer]

	// LargeBufferPool manages 16,384-byte (16 KiB) buffers via indirect indexing.
	LargeBufferPool = IndirectPool[LargeBuffer]

	// HugeBufferPool manages 65,536-byte (64 KiB) buffers via indirect indexing.
	HugeBufferPool = IndirectPool[HugeBuffer]

	// GiantBufferPool manages 262,144-byte (256 KiB) buffers via indirect indexing.
	GiantBufferPool = IndirectPool[GiantBuffer]
)

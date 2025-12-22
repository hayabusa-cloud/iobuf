// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package iobuf

import (
	"math"
	"sync/atomic"
	"unsafe"

	"code.hybscloud.com/iobuf/internal"
	"code.hybscloud.com/iox"
	"code.hybscloud.com/spin"
)

type (
	// PicoBufferBoundedPool implements a bounded MPMC pool for 16-byte buffers.
	PicoBufferBoundedPool = BoundedPool[PicoBuffer]
	// NanoBufferBoundedPool implements a bounded MPMC pool for 64-byte buffers.
	NanoBufferBoundedPool = BoundedPool[NanoBuffer]
	// MicroBufferBoundedPool implements a bounded MPMC pool for 256-byte buffers.
	MicroBufferBoundedPool = BoundedPool[MicroBuffer]
	// SmallBufferBoundedPool implements a bounded MPMC pool for 1 KiB buffers.
	SmallBufferBoundedPool = BoundedPool[SmallBuffer]
	// MediumBufferBoundedPool implements a bounded MPMC pool for 4 KiB buffers.
	MediumBufferBoundedPool = BoundedPool[MediumBuffer]
	// LargeBufferBoundedPool implements a bounded MPMC pool for 16 KiB buffers.
	LargeBufferBoundedPool = BoundedPool[LargeBuffer]
	// HugeBufferBoundedPool implements a bounded MPMC pool for 64 KiB buffers.
	HugeBufferBoundedPool = BoundedPool[HugeBuffer]
	// GiantBufferBoundedPool implements a bounded MPMC pool for 256 KiB buffers.
	GiantBufferBoundedPool = BoundedPool[GiantBuffer]
)

// NewPicoBufferPool creates a new instance of PicoBufferBoundedPool with the specified capacity.
// The capacity must be between 1 and math.MaxUint32 and will be rounded up to the next power of two.
func NewPicoBufferPool(capacity int) *PicoBufferBoundedPool {
	return NewBoundedPool[PicoBuffer](capacity)
}

// NewNanoBufferPool creates a new instance of NanoBufferBoundedPool with the specified capacity.
// The capacity must be between 1 and math.MaxUint32 and will be rounded up to the next power of two.
func NewNanoBufferPool(capacity int) *NanoBufferBoundedPool {
	return NewBoundedPool[NanoBuffer](capacity)
}

// NewMicroBufferPool creates a new instance of MicroBufferBoundedPool with the specified capacity.
// The capacity must be between 1 and math.MaxUint32 and will be rounded up to the next power of two.
func NewMicroBufferPool(capacity int) *MicroBufferBoundedPool {
	return NewBoundedPool[MicroBuffer](capacity)
}

// NewSmallBufferPool creates a new instance of SmallBufferBoundedPool with the specified capacity.
// The capacity must be between 1 and math.MaxUint32 and will be rounded up to the next power of two.
func NewSmallBufferPool(capacity int) *SmallBufferBoundedPool {
	return NewBoundedPool[SmallBuffer](capacity)
}

// NewMediumBufferPool creates a new instance of MediumBufferBoundedPool with the specified capacity.
// The capacity must be between 1 and math.MaxUint32 and will be rounded up to the next power of two.
func NewMediumBufferPool(capacity int) *MediumBufferBoundedPool {
	return NewBoundedPool[MediumBuffer](capacity)
}

// NewLargeBufferPool creates a new instance of LargeBufferBoundedPool with the specified capacity.
// The capacity must be between 1 and math.MaxUint32 and will be rounded up to the next power of two.
func NewLargeBufferPool(capacity int) *LargeBufferBoundedPool {
	return NewBoundedPool[LargeBuffer](capacity)
}

// NewHugeBufferPool creates a new instance of HugeBufferBoundedPool with the specified capacity.
// The capacity must be between 1 and math.MaxUint32 and will be rounded up to the next power of two.
func NewHugeBufferPool(capacity int) *HugeBufferBoundedPool {
	return NewBoundedPool[HugeBuffer](capacity)
}

// NewGiantBufferPool creates a new instance of GiantBufferBoundedPool with the specified capacity.
// The capacity must be between 1 and math.MaxUint32 and will be rounded up to the next power of two.
func NewGiantBufferPool(capacity int) *GiantBufferBoundedPool {
	return NewBoundedPool[GiantBuffer](capacity)
}

// BoundedPoolItem is an interface that represents an item that can be used in a bounded pool.
// Implementing this interface allows an object to be stored and retrieved from a BoundedPool.
type BoundedPoolItem interface{}

// NewBoundedPool creates a new instance of BoundedPool with the specified capacity.
// The capacity must be between 1 and math.MaxUint32 (inclusive).
func NewBoundedPool[ItemType BoundedPoolItem](capacity int) *BoundedPool[ItemType] {
	if capacity < 1 || capacity > math.MaxUint32 {
		panic("capacity must be between 1 and MaxUint32")
	}
	capacity--
	capacity |= capacity >> 1
	capacity |= capacity >> 2
	capacity |= capacity >> 4
	capacity |= capacity >> 8
	capacity |= capacity >> 16
	capacity++

	items := make([]ItemType, 0, capacity)

	remapM := min(internal.CacheLineSize/unsafe.Sizeof(atomic.Uint64{}), uintptr(capacity))
	remapN := max(1, uintptr(capacity)/remapM)
	remapMask := remapN - 1

	ret := BoundedPool[ItemType]{
		items:     items,
		capacity:  uint32(capacity),
		mask:      uint32(capacity - 1),
		remapM:    uint32(remapM),
		remapN:    uint32(remapN),
		remapMask: uint32(remapMask),
		head:      atomic.Uint32{},
		tail:      atomic.Uint32{},

		nonblocking: false,
	}
	return &ret
}

// BoundedPool is a generic type that represents a bounded pool of items of type T.
// The pool has a bounded and fixed capacity and allows items to be retrieved and returned.
// If the pool is empty and the non-blocking mode is not set,
// Get() calls would block until an item is available.
// If the pool is full and the non-blocking mode is not set,
// Put() calls would block until the BoundedPool is no longer full.
// BoundedPool is safe for concurrent use.
// The Get() and Put() methods ensure that at least one of the goroutines makes progress.
// The implementation of BoundedPool is based on the algorithms in the following paper:
//  https://nikitakoval.org/publications/ppopp20-queues.pdf
// Usage:
//
//	pool := NewBoundedPool[ItemType](capacity) creates a new instance of BoundedPool with the specified capacity.
//	pool.Fill(newFunc) initializes and fills the pool with a function to create new items.
//	pool.SetNonblock(nonblocking) enables or disables the non-blocking mode of the pool.
//	pool.Value(indirect) returns the item at the specified indirect index.
//	pool.SetValue(indirect, val) sets the value of the item at the specified indirect index in pool.
//	pool.Get() retrieves an item from the pool and returns its indirect index.
//	pool.Put(indirect) puts the indirect index of an item back into the pool.

type BoundedPool[T BoundedPoolItem] struct {
	_ noCopy

	items      []T
	capacity   uint32
	mask       uint32
	entries    []atomic.Uint64
	remapM     uint32
	remapN     uint32
	remapMask  uint32
	head, tail atomic.Uint32

	nonblocking bool
}

// Fill initializes and fills the BoundedPool with a newFunc function, which is used to create new items.
// Fill Put capacity items with new BoundedPoolItem created by newFunc for each item in the pool.
//
// Example:
//
//	pool := NewBoundedPool[ItemType](capacity)
//	pool.Fill(newFunc)
//
// Parameters:
//
//	newFunc - a function that returns an instance of an item to be added to the pool.
func (pool *BoundedPool[T]) Fill(newFunc func() T) {
	for range pool.capacity {
		pool.items = append(pool.items, newFunc())
	}
	pool.entries = make([]atomic.Uint64, pool.capacity)
	for i := range pool.capacity {
		pool.entries[i].Store(uint64(i))
	}
	pool.tail.Store(pool.capacity)
}

// SetNonblock enables or disables the non-blocking mode of the pool.
// When nonblocking is set to true, Get() and Put() calls will not block and return immediately.
// When nonblocking is set to false, Get() calls will block until an item is available,
// and Put() calls will block until the pool is no longer full.
//
// Example:
//
//	pool := NewBoundedPool[ItemType](capacity)
//	pool.SetNonblock(true)
//
// Parameters:
//
//	nonblocking - determines whether the pool operates in non-blocking mode (true) or blocking mode (false).
func (pool *BoundedPool[T]) SetNonblock(nonblocking bool) {
	pool.nonblocking = nonblocking
}

// Value returns the item at the specified indirect index.
// The given indirect index must not be marked as empty and must be within the valid range.
func (pool *BoundedPool[T]) Value(indirect int) T {
	if len(pool.items) != int(pool.capacity) {
		panic("must Fill the pool before using it")
	}
	if indirect&boundedPoolEntryEmpty == boundedPoolEntryEmpty {
		panic("invalid bounded pool indirect")
	}
	if indirect < 0 || indirect >= int(pool.capacity) {
		panic("invalid bounded pool indirect")
	}

	return pool.items[indirect]
}

// SetValue sets the value of the item at the specified indirect index in the BoundedPool.
// The given indirect index must not be marked as empty and must be within the valid range.
func (pool *BoundedPool[T]) SetValue(indirect int, value T) {
	if len(pool.items) != int(pool.capacity) {
		panic("must Fill the pool before using it")
	}
	if indirect&boundedPoolEntryEmpty == boundedPoolEntryEmpty {
		panic("invalid bounded pool indirect")
	}
	if indirect < 0 || indirect >= int(pool.capacity) {
		panic("invalid bounded pool indirect")
	}

	pool.items[indirect] = value
}

// Get retrieves an item from the pool and returns its indirect index.
// If an item is available, its indirect index and a nil error are returned.
// Returns iox.ErrWouldBlock if the pool is empty and nonblocking mode is set.
//
// In blocking mode, Get uses adaptive waiting (iox.Backoff) when the
// pool is empty. This acknowledges that buffer exhaustion is an external I/O
// event—buffers are released when the kernel/network finishes processing—
// requiring OS-level sleep rather than hardware-level spin.
func (pool *BoundedPool[T]) Get() (indirect int, err error) {
	if len(pool.items) != int(pool.capacity) {
		panic("must Fill the pool before using it")
	}
	var aw iox.Backoff
	for {
		entry, err := pool.tryGet()
		if err == nil {
			return int(entry & uint64(pool.mask)), nil
		}
		if err == iox.ErrWouldBlock {
			if pool.nonblocking {
				return boundedPoolEntryEmpty, err
			}
			// Buffer exhaustion: external I/O scale event.
			// Use adaptive waiting to yield CPU while waiting for
			// network/disk completion to release buffers.
			aw.Wait()
			continue
		}
		return boundedPoolEntryEmpty, err
	}
}

// Put puts the indirect index of an item back into the BoundedPool.
// It tries to put the given indirect index into the pool and returns
// nil error if successful. If the BoundedPool is currently full, it
// would block until the item can be put into the pool or return
// iox.ErrWouldBlock if the pool is nonblocking.
//
// In blocking mode, Put uses adaptive waiting (iox.Backoff) when the
// pool is full. This acknowledges that pool capacity is freed by external
// consumers completing their I/O operations.
func (pool *BoundedPool[T]) Put(indirect int) error {
	if len(pool.items) != int(pool.capacity) {
		panic("must Fill the pool before using it")
	}
	entry := uint64(indirect)
	var aw iox.Backoff
	for {
		err := pool.tryPut(entry)
		if err == nil {
			return nil
		}
		if err == iox.ErrWouldBlock {
			if pool.nonblocking {
				return err
			}
			// Pool full: external consumer scale event.
			// Use adaptive waiting to yield CPU while waiting for
			// consumers to complete their operations.
			aw.Wait()
			continue
		}
		return err
	}
}

// Cap returns the capacity of the BoundedPool
func (pool *BoundedPool[T]) Cap() int {
	return int(pool.capacity)
}

const (
	boundedPoolEntryEmpty    = 1 << 62
	boundedPoolEntryTurnMask = boundedPoolEntryEmpty>>32 - 1
)

func (pool *BoundedPool[T]) tryGet() (entry uint64, err error) {
	sw := spin.Wait{}
	for {
		h, t := pool.head.Load(), pool.tail.Load()
		hi := pool.remap(h & pool.mask)
		e := pool.entries[hi].Load()

		if h != pool.head.Load() {
			sw.Once()
			continue
		}

		if h == t {
			return boundedPoolEntryEmpty, iox.ErrWouldBlock
		}

		nextTurn := (h/pool.capacity + 1) & boundedPoolEntryTurnMask
		if e == pool.empty(nextTurn) {
			pool.head.CompareAndSwap(h, h+1)
			sw.Once()
			continue
		}
		ok := pool.entries[hi].CompareAndSwap(e, pool.empty(nextTurn))
		pool.head.CompareAndSwap(h, h+1)
		if ok {
			return e, nil
		}
		sw.Once()
	}
}

func (pool *BoundedPool[T]) tryPut(e uint64) error {
	sw := spin.Wait{}
	for {
		h, t := pool.head.Load(), pool.tail.Load()
		if t != pool.tail.Load() {
			sw.Once()
			continue
		}
		if t == h+pool.capacity {
			return iox.ErrWouldBlock
		}
		turn, ti := (t/pool.capacity)&boundedPoolEntryTurnMask, pool.remap(t)
		ok := pool.entries[ti].CompareAndSwap(pool.empty(turn), e)
		pool.tail.CompareAndSwap(t, t+1)
		if ok {
			return nil
		}
		sw.Once()
	}
}

func (pool *BoundedPool[T]) remap(cursor uint32) int {
	p, q := cursor/pool.remapN, cursor&pool.remapMask
	return int(q*pool.remapM + p%pool.remapM)
}

func (pool *BoundedPool[T]) empty(turn uint32) uint64 {
	return boundedPoolEntryEmpty | uint64(turn&boundedPoolEntryTurnMask)
}

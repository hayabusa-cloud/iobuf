// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package iobuf

import (
	"sync/atomic"
	"unsafe"

	"code.hybscloud.com/iobuf/internal"
	"code.hybscloud.com/iox"
	"code.hybscloud.com/spin"
)

type (
	// PicoBufferBoundedPool implements a bounded MPMC pool for 32-byte buffers.
	PicoBufferBoundedPool = BoundedPool[PicoBuffer]
	// NanoBufferBoundedPool implements a bounded MPMC pool for 128-byte buffers.
	NanoBufferBoundedPool = BoundedPool[NanoBuffer]
	// MicroBufferBoundedPool implements a bounded MPMC pool for 512-byte buffers.
	MicroBufferBoundedPool = BoundedPool[MicroBuffer]
	// SmallBufferBoundedPool implements a bounded MPMC pool for 2 KiB buffers.
	SmallBufferBoundedPool = BoundedPool[SmallBuffer]
	// MediumBufferBoundedPool implements a bounded MPMC pool for 8 KiB buffers.
	MediumBufferBoundedPool = BoundedPool[MediumBuffer]
	// BigBufferBoundedPool implements a bounded MPMC pool for 32 KiB buffers.
	BigBufferBoundedPool = BoundedPool[BigBuffer]
	// LargeBufferBoundedPool implements a bounded MPMC pool for 128 KiB buffers.
	LargeBufferBoundedPool = BoundedPool[LargeBuffer]
	// GreatBufferBoundedPool implements a bounded MPMC pool for 512 KiB buffers.
	GreatBufferBoundedPool = BoundedPool[GreatBuffer]
	// HugeBufferBoundedPool implements a bounded MPMC pool for 2 MiB buffers.
	HugeBufferBoundedPool = BoundedPool[HugeBuffer]
	// VastBufferBoundedPool implements a bounded MPMC pool for 8 MiB buffers.
	VastBufferBoundedPool = BoundedPool[VastBuffer]
	// GiantBufferBoundedPool implements a bounded MPMC pool for 32 MiB buffers.
	GiantBufferBoundedPool = BoundedPool[GiantBuffer]
	// TitanBufferBoundedPool implements a bounded MPMC pool for 128 MiB buffers.
	TitanBufferBoundedPool = BoundedPool[TitanBuffer]
)

// NewPicoBufferPool creates a new instance of PicoBufferBoundedPool with the specified capacity.
// The capacity must be at least 1 and must round to no more than 1<<31.
func NewPicoBufferPool(capacity int) *PicoBufferBoundedPool {
	return NewBoundedPool[PicoBuffer](capacity)
}

// NewNanoBufferPool creates a new instance of NanoBufferBoundedPool with the specified capacity.
// The capacity must be at least 1 and must round to no more than 1<<31.
func NewNanoBufferPool(capacity int) *NanoBufferBoundedPool {
	return NewBoundedPool[NanoBuffer](capacity)
}

// NewMicroBufferPool creates a new instance of MicroBufferBoundedPool with the specified capacity.
// The capacity must be at least 1 and must round to no more than 1<<31.
func NewMicroBufferPool(capacity int) *MicroBufferBoundedPool {
	return NewBoundedPool[MicroBuffer](capacity)
}

// NewSmallBufferPool creates a new instance of SmallBufferBoundedPool with the specified capacity.
// The capacity must be at least 1 and must round to no more than 1<<31.
func NewSmallBufferPool(capacity int) *SmallBufferBoundedPool {
	return NewBoundedPool[SmallBuffer](capacity)
}

// NewMediumBufferPool creates a new instance of MediumBufferBoundedPool with the specified capacity.
// The capacity must be at least 1 and must round to no more than 1<<31.
func NewMediumBufferPool(capacity int) *MediumBufferBoundedPool {
	return NewBoundedPool[MediumBuffer](capacity)
}

// NewBigBufferPool creates a new instance of BigBufferBoundedPool with the specified capacity.
// The capacity must be at least 1 and must round to no more than 1<<31.
func NewBigBufferPool(capacity int) *BigBufferBoundedPool {
	return NewBoundedPool[BigBuffer](capacity)
}

// NewLargeBufferPool creates a new instance of LargeBufferBoundedPool with the specified capacity.
// The capacity must be at least 1 and must round to no more than 1<<31.
func NewLargeBufferPool(capacity int) *LargeBufferBoundedPool {
	return NewBoundedPool[LargeBuffer](capacity)
}

// NewGreatBufferPool creates a new instance of GreatBufferBoundedPool with the specified capacity.
// The capacity must be at least 1 and must round to no more than 1<<31.
func NewGreatBufferPool(capacity int) *GreatBufferBoundedPool {
	return NewBoundedPool[GreatBuffer](capacity)
}

// NewHugeBufferPool creates a new instance of HugeBufferBoundedPool with the specified capacity.
// The capacity must be at least 1 and must round to no more than 1<<31.
func NewHugeBufferPool(capacity int) *HugeBufferBoundedPool {
	return NewBoundedPool[HugeBuffer](capacity)
}

// NewVastBufferPool creates a new instance of VastBufferBoundedPool with the specified capacity.
// The capacity must be at least 1 and must round to no more than 1<<31.
func NewVastBufferPool(capacity int) *VastBufferBoundedPool {
	return NewBoundedPool[VastBuffer](capacity)
}

// NewGiantBufferPool creates a new instance of GiantBufferBoundedPool with the specified capacity.
// The capacity must be at least 1 and must round to no more than 1<<31.
func NewGiantBufferPool(capacity int) *GiantBufferBoundedPool {
	return NewBoundedPool[GiantBuffer](capacity)
}

// NewTitanBufferPool creates a new instance of TitanBufferBoundedPool with the specified capacity.
// The capacity must be at least 1 and must round to no more than 1<<31.
func NewTitanBufferPool(capacity int) *TitanBufferBoundedPool {
	return NewBoundedPool[TitanBuffer](capacity)
}

// BoundedPoolItem is a type constraint for items stored in a BoundedPool.
//
// Any type can satisfy this constraint. The alias exists to make the
// generic type parameter explicit and to allow future extension.
type BoundedPoolItem = any

// NewBoundedPool creates a lock-free bounded pool with the specified capacity.
//
// The capacity is rounded up to the next power of two for efficient index
// calculation. The actual capacity can be retrieved via Cap().
//
// Panics if capacity < 1 or if rounding would exceed 1<<31.
//
// After creation, Fill must be called before Get/Put operations.
func NewBoundedPool[ItemType BoundedPoolItem](capacity int) *BoundedPool[ItemType] {
	if capacity < 1 {
		panic("capacity must be at least 1")
	}
	capacity--
	capacity |= capacity >> 1
	capacity |= capacity >> 2
	capacity |= capacity >> 4
	capacity |= capacity >> 8
	capacity |= capacity >> 16
	capacity++
	if capacity > boundedPoolMaxCapacity {
		panic("capacity must round to no more than 1<<31")
	}

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

// BoundedPool is a fixed-capacity pool of items addressed by index.
// After initialization, multiple goroutines can acquire and return indices
// concurrently. Each caller controls its slot until it returns the index.
//
// The index queue is based on [Restricted Memory-Friendly Lock-Free Bounded Queues].
//
// [Restricted Memory-Friendly Lock-Free Bounded Queues]: https://nikitakoval.org/publications/ppopp20-queues.pdf
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

// Fill initializes each slot with one call to newFunc.
// Call Fill once before sharing the pool or acquiring its first slot.
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

// SetNonblock selects how Get handles an empty pool and Put handles a full pool.
// The default is to wait. In nonblocking mode both methods return
// [iox.ErrWouldBlock] in those cases.
// Both modes retry contended queue updates with [spin.Wait].
// Configure the mode before sharing the pool.
func (pool *BoundedPool[T]) SetNonblock(nonblocking bool) {
	pool.nonblocking = nonblocking
}

// Value returns the item at the caller's acquired index by value. An array is
// copied. A slice shares its backing array with the item in the pool.
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

// SetValue replaces the item at the caller's acquired index.
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

// Get acquires a slot and returns its index. When the pool is empty, Get waits
// with [iox.Backoff] in blocking mode and returns [iox.ErrWouldBlock] in
// nonblocking mode. Use the index only when Get succeeds.
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
		if pool.nonblocking {
			return boundedPoolEntryEmpty, err
		}
		aw.Wait()
	}
}

// Put returns an acquired index. Return it once, after all application and
// kernel access to the slot has ended. When the pool is full, Put waits with
// [iox.Backoff] in blocking mode and returns [iox.ErrWouldBlock] in nonblocking
// mode. An unsuccessful Put leaves the slot with its caller.
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
		if pool.nonblocking {
			return err
		}
		aw.Wait()
	}
}

// Cap returns the actual capacity of the BoundedPool.
//
// This may be larger than the requested capacity due to power-of-two rounding.
func (pool *BoundedPool[T]) Cap() int {
	return int(pool.capacity)
}

// Internal constants for the lock-free FIFO algorithm.
// Entry format: [turn:30][reserved:2][empty:1][index:31]
const (
	boundedPoolMaxCapacity   = 1 << 31                       // Largest representable rounded capacity
	boundedPoolEntryEmpty    = 1 << 62                       // Marks slot as empty
	boundedPoolEntryTurnMask = boundedPoolEntryEmpty>>32 - 1 // Mask for turn counter
)

// tryGet attempts a single non-blocking dequeue from the pool.
// Returns the entry value and nil on success, or boundedPoolEntryEmpty
// and ErrWouldBlock if the pool is empty.
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

// tryPut attempts a single non-blocking enqueue into the pool.
// Returns nil on success, or ErrWouldBlock if the pool is full.
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

// remap converts a logical cursor position to a physical array index.
// This remapping improves cache locality by distributing adjacent logical
// positions across different cache lines.
func (pool *BoundedPool[T]) remap(cursor uint32) int {
	p, q := cursor/pool.remapN, cursor&pool.remapMask
	return int(q*pool.remapM + p%pool.remapM)
}

// empty creates an empty marker with the given turn counter.
// The turn counter prevents ABA problems by ensuring entries are unique
// across different enqueue/dequeue cycles.
func (pool *BoundedPool[T]) empty(turn uint32) uint64 {
	return boundedPoolEntryEmpty | uint64(turn&boundedPoolEntryTurnMask)
}

// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package iobuf_test

import (
	"sync"
	"testing"

	"code.hybscloud.com/iobuf"
	"code.hybscloud.com/iox"
	"code.hybscloud.com/spin"
)

func TestBoundedPool_BasicGetPut(t *testing.T) {
	const capacity = 16
	pool := iobuf.NewBoundedPool[int](capacity)

	// Fill the pool with values
	counter := 0
	pool.Fill(func() int {
		v := counter * 10
		counter++
		return v
	})

	// Get all items
	indices := make([]int, capacity)
	for i := range capacity {
		idx, err := pool.Get()
		if err != nil {
			t.Fatalf("Get() failed at iteration %d: %v", i, err)
		}
		indices[i] = idx
	}

	// Put all items back
	for _, idx := range indices {
		err := pool.Put(idx)
		if err != nil {
			t.Fatalf("Put(%d) failed: %v", idx, err)
		}
	}

	// Verify we can get them again
	for i := range capacity {
		_, err := pool.Get()
		if err != nil {
			t.Fatalf("Second Get() failed at iteration %d: %v", i, err)
		}
	}
}

func TestBoundedPool_NonblockingEmpty(t *testing.T) {
	const capacity = 4
	pool := iobuf.NewBoundedPool[int](capacity)
	pool.SetNonblock(true)

	pool.Fill(func() int { return 0 })

	// Drain the pool
	for range capacity {
		_, err := pool.Get()
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
	}

	// Next Get should return ErrWouldBlock
	_, err := pool.Get()
	if err != iox.ErrWouldBlock {
		t.Errorf("expected iox.ErrWouldBlock, got %v", err)
	}
}

func TestBoundedPool_NonblockingFull(t *testing.T) {
	const capacity = 4
	pool := iobuf.NewBoundedPool[int](capacity)
	pool.SetNonblock(true)

	pool.Fill(func() int { return 0 })

	// Pool is full, Put should return ErrWouldBlock
	err := pool.Put(0)
	if err != iox.ErrWouldBlock {
		t.Errorf("expected iox.ErrWouldBlock on full pool, got %v", err)
	}
}

func TestBoundedPool_Concurrent(t *testing.T) {
	const capacity = 64
	const goroutines = 16
	const iterations = 2000

	pool := iobuf.NewBoundedPool[int](capacity)
	pool.Fill(func() int { return 0 })

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for i := range iterations {
				idx, err := pool.Get()
				if err != nil {
					t.Errorf("goroutine %d iteration %d: Get() failed: %v", id, i, err)
					return
				}
				// Simulate some work
				_ = pool.Value(idx)
				spin.Yield()
				err = pool.Put(idx)
				if err != nil {
					t.Errorf("goroutine %d iteration %d: Put() failed: %v", id, i, err)
					return
				}
			}
		}(g)
	}

	wg.Wait()
}

func TestBoundedPool_HighContention(t *testing.T) {
	// High contention test with many goroutines on small pool
	const capacity = 8
	const goroutines = 16
	const iterations = 2000

	pool := iobuf.NewBoundedPool[int](capacity)
	pool.Fill(func() int { return 0 })

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()
			for range iterations {
				idx, err := pool.Get()
				if err != nil {
					spin.Yield()
					continue
				}
				spin.Yield()
				_ = pool.Put(idx)
			}
		}()
	}

	wg.Wait()
}

func TestBoundedPool_Cap(t *testing.T) {
	const capacity = 32
	pool := iobuf.NewBoundedPool[int](capacity)
	if pool.Cap() != capacity {
		t.Errorf("Cap() = %d, want %d", pool.Cap(), capacity)
	}
}

func TestBoundedPool_Value(t *testing.T) {
	const capacity = 8
	pool := iobuf.NewBoundedPool[string](capacity)

	pool.Fill(func() string { return "item" })

	// Get an item and modify it
	idx, err := pool.Get()
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	pool.SetValue(idx, "modified")
	if pool.Value(idx) != "modified" {
		t.Errorf("Value(%d) = %q, want %q", idx, pool.Value(idx), "modified")
	}

	// Put it back
	err = pool.Put(idx)
	if err != nil {
		t.Fatalf("Put() failed: %v", err)
	}
}

func TestNewBoundedPool_InvalidCapacity(t *testing.T) {
	t.Run("zero capacity", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("NewBoundedPool(0) did not panic")
			}
		}()
		_ = iobuf.NewBoundedPool[int](0)
	})

	t.Run("negative capacity", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("NewBoundedPool(-1) did not panic")
			}
		}()
		_ = iobuf.NewBoundedPool[int](-1)
	})
}

func TestBoundedPool_Value_PanicUnfilled(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Value() on unfilled pool did not panic")
		}
	}()
	pool := iobuf.NewBoundedPool[int](8)
	_ = pool.Value(0)
}

func TestBoundedPool_SetValue_PanicUnfilled(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("SetValue() on unfilled pool did not panic")
		}
	}()
	pool := iobuf.NewBoundedPool[int](8)
	pool.SetValue(0, 42)
}

func TestBoundedPool_Value_PanicInvalidIndirect(t *testing.T) {
	pool := iobuf.NewBoundedPool[int](8)
	pool.Fill(func() int { return 0 })

	t.Run("negative index", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Value(-1) did not panic")
			}
		}()
		_ = pool.Value(-1)
	})

	t.Run("out of range index", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Value(capacity) did not panic")
			}
		}()
		_ = pool.Value(pool.Cap())
	})
}

func TestBoundedPool_SetValue_PanicInvalidIndirect(t *testing.T) {
	pool := iobuf.NewBoundedPool[int](8)
	pool.Fill(func() int { return 0 })

	t.Run("negative index", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("SetValue(-1, v) did not panic")
			}
		}()
		pool.SetValue(-1, 42)
	})

	t.Run("out of range index", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("SetValue(capacity, v) did not panic")
			}
		}()
		pool.SetValue(pool.Cap(), 42)
	})
}

func TestBoundedPool_Get_PanicUnfilled(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Get() on unfilled pool did not panic")
		}
	}()
	pool := iobuf.NewBoundedPool[int](8)
	_, _ = pool.Get()
}

func TestBoundedPool_Put_PanicUnfilled(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Put() on unfilled pool did not panic")
		}
	}()
	pool := iobuf.NewBoundedPool[int](8)
	_ = pool.Put(0)
}

func TestBoundedPool_BlockingGet(t *testing.T) {
	const capacity = 4
	pool := iobuf.NewBoundedPool[int](capacity)
	pool.Fill(func() int { return 0 })

	// Drain the pool
	indices := make([]int, capacity)
	for i := range capacity {
		idx, err := pool.Get()
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		indices[i] = idx
	}

	// Start a goroutine that will Put after a short delay
	done := make(chan struct{})
	go func() {
		defer close(done)
		// Small delay to ensure Get() enters blocking state
		for range 1000 {
			spin.Yield()
		}
		// Return one item to unblock the waiting Get
		_ = pool.Put(indices[0])
	}()

	// This Get should block until the Put above completes
	idx, err := pool.Get()
	if err != nil {
		t.Fatalf("blocking Get() failed: %v", err)
	}
	_ = idx

	<-done
}

func TestBoundedPool_BlockingPut(t *testing.T) {
	const capacity = 4
	pool := iobuf.NewBoundedPool[int](capacity)
	pool.Fill(func() int { return 0 })

	// Pool is already full after Fill

	// Start a goroutine that will Get after a short delay
	done := make(chan struct{})
	go func() {
		defer close(done)
		// Small delay to ensure Put() enters blocking state
		for range 1000 {
			spin.Yield()
		}
		// Get one item to unblock the waiting Put
		_, _ = pool.Get()
	}()

	// This Put should block until the Get above completes
	err := pool.Put(0)
	if err != nil {
		t.Fatalf("blocking Put() failed: %v", err)
	}

	<-done
}

func TestNewTierBufferPools(t *testing.T) {
	const capacity = 16

	t.Run("NewPicoBufferPool", func(t *testing.T) {
		pool := iobuf.NewPicoBufferPool(capacity)
		if pool.Cap() != capacity {
			t.Errorf("NewPicoBufferPool capacity = %d, want %d", pool.Cap(), capacity)
		}
		pool.Fill(iobuf.NewPicoBuffer)
		idx, err := pool.Get()
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		buf := pool.Value(idx)
		if len(buf) != iobuf.BufferSizePico {
			t.Errorf("buffer size = %d, want %d", len(buf), iobuf.BufferSizePico)
		}
	})

	t.Run("NewNanoBufferPool", func(t *testing.T) {
		pool := iobuf.NewNanoBufferPool(capacity)
		if pool.Cap() != capacity {
			t.Errorf("NewNanoBufferPool capacity = %d, want %d", pool.Cap(), capacity)
		}
		pool.Fill(iobuf.NewNanoBuffer)
		idx, err := pool.Get()
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		buf := pool.Value(idx)
		if len(buf) != iobuf.BufferSizeNano {
			t.Errorf("buffer size = %d, want %d", len(buf), iobuf.BufferSizeNano)
		}
	})

	t.Run("NewMicroBufferPool", func(t *testing.T) {
		pool := iobuf.NewMicroBufferPool(capacity)
		if pool.Cap() != capacity {
			t.Errorf("NewMicroBufferPool capacity = %d, want %d", pool.Cap(), capacity)
		}
		pool.Fill(iobuf.NewMicroBuffer)
		idx, err := pool.Get()
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		buf := pool.Value(idx)
		if len(buf) != iobuf.BufferSizeMicro {
			t.Errorf("buffer size = %d, want %d", len(buf), iobuf.BufferSizeMicro)
		}
	})

	t.Run("NewSmallBufferPool", func(t *testing.T) {
		pool := iobuf.NewSmallBufferPool(capacity)
		if pool.Cap() != capacity {
			t.Errorf("NewSmallBufferPool capacity = %d, want %d", pool.Cap(), capacity)
		}
		pool.Fill(iobuf.NewSmallBuffer)
		idx, err := pool.Get()
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		buf := pool.Value(idx)
		if len(buf) != iobuf.BufferSizeSmall {
			t.Errorf("buffer size = %d, want %d", len(buf), iobuf.BufferSizeSmall)
		}
	})

	t.Run("NewMediumBufferPool", func(t *testing.T) {
		pool := iobuf.NewMediumBufferPool(capacity)
		if pool.Cap() != capacity {
			t.Errorf("NewMediumBufferPool capacity = %d, want %d", pool.Cap(), capacity)
		}
		pool.Fill(iobuf.NewMediumBuffer)
		idx, err := pool.Get()
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		buf := pool.Value(idx)
		if len(buf) != iobuf.BufferSizeMedium {
			t.Errorf("buffer size = %d, want %d", len(buf), iobuf.BufferSizeMedium)
		}
	})

	t.Run("NewBigBufferPool", func(t *testing.T) {
		pool := iobuf.NewBigBufferPool(capacity)
		if pool.Cap() != capacity {
			t.Errorf("NewBigBufferPool capacity = %d, want %d", pool.Cap(), capacity)
		}
		pool.Fill(iobuf.NewBigBuffer)
		idx, err := pool.Get()
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		buf := pool.Value(idx)
		if len(buf) != iobuf.BufferSizeBig {
			t.Errorf("buffer size = %d, want %d", len(buf), iobuf.BufferSizeBig)
		}
	})

	t.Run("NewLargeBufferPool", func(t *testing.T) {
		pool := iobuf.NewLargeBufferPool(capacity)
		if pool.Cap() != capacity {
			t.Errorf("NewLargeBufferPool capacity = %d, want %d", pool.Cap(), capacity)
		}
		pool.Fill(iobuf.NewLargeBuffer)
		idx, err := pool.Get()
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		buf := pool.Value(idx)
		if len(buf) != iobuf.BufferSizeLarge {
			t.Errorf("buffer size = %d, want %d", len(buf), iobuf.BufferSizeLarge)
		}
	})

	t.Run("NewGreatBufferPool", func(t *testing.T) {
		pool := iobuf.NewGreatBufferPool(capacity)
		if pool.Cap() != capacity {
			t.Errorf("NewGreatBufferPool capacity = %d, want %d", pool.Cap(), capacity)
		}
		pool.Fill(iobuf.NewGreatBuffer)
		idx, err := pool.Get()
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		buf := pool.Value(idx)
		if len(buf) != iobuf.BufferSizeGreat {
			t.Errorf("buffer size = %d, want %d", len(buf), iobuf.BufferSizeGreat)
		}
	})

	t.Run("NewHugeBufferPool", func(t *testing.T) {
		pool := iobuf.NewHugeBufferPool(capacity)
		if pool.Cap() != capacity {
			t.Errorf("NewHugeBufferPool capacity = %d, want %d", pool.Cap(), capacity)
		}
		pool.Fill(iobuf.NewHugeBuffer)
		idx, err := pool.Get()
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		buf := pool.Value(idx)
		if len(buf) != iobuf.BufferSizeHuge {
			t.Errorf("buffer size = %d, want %d", len(buf), iobuf.BufferSizeHuge)
		}
	})

	t.Run("NewVastBufferPool", func(t *testing.T) {
		const smallCap = 2 // Use small capacity for large buffers
		pool := iobuf.NewVastBufferPool(smallCap)
		if pool.Cap() != smallCap {
			t.Errorf("NewVastBufferPool capacity = %d, want %d", pool.Cap(), smallCap)
		}
		pool.Fill(iobuf.NewVastBuffer)
		idx, err := pool.Get()
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		buf := pool.Value(idx)
		if len(buf) != iobuf.BufferSizeVast {
			t.Errorf("buffer size = %d, want %d", len(buf), iobuf.BufferSizeVast)
		}
	})

	t.Run("NewGiantBufferPool", func(t *testing.T) {
		const smallCap = 2 // Use small capacity for large buffers
		pool := iobuf.NewGiantBufferPool(smallCap)
		if pool.Cap() != smallCap {
			t.Errorf("NewGiantBufferPool capacity = %d, want %d", pool.Cap(), smallCap)
		}
		pool.Fill(iobuf.NewGiantBuffer)
		idx, err := pool.Get()
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		buf := pool.Value(idx)
		if len(buf) != iobuf.BufferSizeGiant {
			t.Errorf("buffer size = %d, want %d", len(buf), iobuf.BufferSizeGiant)
		}
	})

	t.Run("NewTitanBufferPool", func(t *testing.T) {
		if raceEnabled {
			t.Skip("TitanBuffer (128 MiB) skipped in race mode due to stack overhead")
		}
		const smallCap = 1 // Use minimal capacity for large buffers
		pool := iobuf.NewTitanBufferPool(smallCap)
		if pool.Cap() != smallCap {
			t.Errorf("NewTitanBufferPool capacity = %d, want %d", pool.Cap(), smallCap)
		}
		pool.Fill(iobuf.NewTitanBuffer)
		idx, err := pool.Get()
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		buf := pool.Value(idx)
		if len(buf) != iobuf.BufferSizeTitan {
			t.Errorf("buffer size = %d, want %d", len(buf), iobuf.BufferSizeTitan)
		}
	})
}

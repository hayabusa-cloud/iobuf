// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package iobuf_test

import (
	"testing"

	"code.hybscloud.com/iobuf"
	"code.hybscloud.com/iox"
	"code.hybscloud.com/spin"
)

// Pool benchmarks

func BenchmarkSmallBufferPool_GetPut(b *testing.B) {
	pool := iobuf.NewSmallBufferPool(1024)
	pool.Fill(iobuf.NewSmallBuffer)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				b.Fatal(err)
			}
			// Simulate I/O latency
			spin.Yield()
			_ = pool.Put(idx)
		}
	})
}

func BenchmarkMediumBufferPool_GetPut(b *testing.B) {
	pool := iobuf.NewMediumBufferPool(1024)
	pool.Fill(iobuf.NewMediumBuffer)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				b.Fatal(err)
			}
			// Simulate I/O latency
			spin.Yield()
			_ = pool.Put(idx)
		}
	})
}

func BenchmarkLargeBufferPool_GetPut(b *testing.B) {
	pool := iobuf.NewLargeBufferPool(1024)
	pool.Fill(iobuf.NewLargeBuffer)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				b.Fatal(err)
			}
			// Simulate I/O latency
			spin.Yield()
			_ = pool.Put(idx)
		}
	})
}

// Memory allocation benchmarks

func BenchmarkAlignedMemBlock(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = iobuf.AlignedMemBlock()
	}
}

func BenchmarkAlignedMem_4K(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = iobuf.AlignedMem(4096, iobuf.PageSize)
	}
}

func BenchmarkAlignedMem_64K(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = iobuf.AlignedMem(65536, iobuf.PageSize)
	}
}

func BenchmarkAlignedMemBlocks_16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = iobuf.AlignedMemBlocks(16, iobuf.PageSize)
	}
}

// IoVec benchmarks

func BenchmarkIoVecFromSmallBuffers_8(b *testing.B) {
	buffers := make([]iobuf.SmallBuffer, 8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = iobuf.IoVecFromSmallBuffers(buffers)
	}
}

func BenchmarkIoVecFromSmallBuffers_64(b *testing.B) {
	buffers := make([]iobuf.SmallBuffer, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = iobuf.IoVecFromSmallBuffers(buffers)
	}
}

func BenchmarkIoVecFromLargeBuffers_8(b *testing.B) {
	buffers := make([]iobuf.LargeBuffer, 8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = iobuf.IoVecFromLargeBuffers(buffers)
	}
}

func BenchmarkIoVecFromBytesSlice_8(b *testing.B) {
	slices := make([][]byte, 8)
	for i := range slices {
		slices[i] = make([]byte, 256)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = iobuf.IoVecFromBytesSlice(slices)
	}
}

func BenchmarkIoVecAddrLen(b *testing.B) {
	buffers := make([]iobuf.SmallBuffer, 8)
	iovecs := iobuf.IoVecFromSmallBuffers(buffers)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = iobuf.IoVecAddrLen(iovecs)
	}
}

// Buffer value access benchmarks

func BenchmarkPool_Value(b *testing.B) {
	pool := iobuf.NewSmallBufferPool(1024)
	pool.Fill(iobuf.NewSmallBuffer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pool.Value(i % 1024)
	}
}

func BenchmarkPool_SetValue(b *testing.B) {
	pool := iobuf.NewSmallBufferPool(1024)
	pool.Fill(iobuf.NewSmallBuffer)
	buf := iobuf.NewSmallBuffer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.SetValue(i%1024, buf)
	}
}

// High-contention benchmarks demonstrating Backoff behavior
//
// These benchmarks simulate buffer exhaustion scenarios where multiple goroutines
// compete for a small pool. When the pool is empty, Get() uses iox.Backoff
// (linear block-backoff with jitter) to wait for buffer release, acknowledging that
// buffer availability is an external I/O event (network/disk completion).

func BenchmarkPool_HighContention_SmallPool(b *testing.B) {
	// Small pool (16 buffers) with high parallelism creates contention
	// This triggers the Backoff when pool is temporarily exhausted
	pool := iobuf.NewSmallBufferPool(16)
	pool.Fill(iobuf.NewSmallBuffer)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var ba iox.Backoff
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				b.Fatal(err)
			}
			// Simulate brief I/O work
			ba.Wait()
			_ = pool.Put(idx)
		}
	})
}

func BenchmarkPool_HighContention_TinyPool(b *testing.B) {
	// Tiny pool (4 buffers) creates extreme contention
	// Backoff will engage frequently with linear progression
	pool := iobuf.NewSmallBufferPool(4)
	pool.Fill(iobuf.NewSmallBuffer)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				b.Fatal(err)
			}
			// Simulate I/O latency
			spin.Yield()
			_ = pool.Put(idx)
		}
	})
}

func BenchmarkPool_Contention_MediumBuffer(b *testing.B) {
	// Medium buffers with moderate contention
	pool := iobuf.NewMediumBufferPool(32)
	pool.Fill(iobuf.NewMediumBuffer)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				b.Fatal(err)
			}
			spin.Yield()
			_ = pool.Put(idx)
		}
	})
}

func BenchmarkPool_Contention_LargeBuffer(b *testing.B) {
	// Large buffers with moderate contention
	pool := iobuf.NewLargeBufferPool(32)
	pool.Fill(iobuf.NewLargeBuffer)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				b.Fatal(err)
			}
			spin.Yield()
			_ = pool.Put(idx)
		}
	})
}

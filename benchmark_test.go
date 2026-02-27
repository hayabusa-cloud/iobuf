// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package iobuf_test

import (
	"testing"

	"code.hybscloud.com/iobuf"
)

// Benchmark Design Notes
//
// Parallel pool benchmarks use nonblocking mode with skip-on-failure pattern.
// Errors are rare with properly-sized pools (1024+ buffers for 10 goroutines).
// CAS contention is handled internally by spin.Wait{} in tryGet()/tryPut().

// Pool benchmarks

func BenchmarkSmallBufferPool_GetPut(b *testing.B) {
	pool := iobuf.NewSmallBufferPool(1024)
	pool.Fill(iobuf.NewSmallBuffer)
	pool.SetNonblock(true)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				continue
			}
			_ = pool.Put(idx)
		}
	})
}

func BenchmarkMediumBufferPool_GetPut(b *testing.B) {
	pool := iobuf.NewMediumBufferPool(1024)
	pool.Fill(iobuf.NewMediumBuffer)
	pool.SetNonblock(true)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				continue
			}
			_ = pool.Put(idx)
		}
	})
}

func BenchmarkLargeBufferPool_GetPut(b *testing.B) {
	pool := iobuf.NewLargeBufferPool(1024)
	pool.Fill(iobuf.NewLargeBuffer)
	pool.SetNonblock(true)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				continue
			}
			_ = pool.Put(idx)
		}
	})
}

// Memory allocation benchmarks

func BenchmarkAlignedMemBlock(b *testing.B) {
	for range b.N {
		_ = iobuf.AlignedMemBlock()
	}
}

func BenchmarkAlignedMem_4K(b *testing.B) {
	for range b.N {
		_ = iobuf.AlignedMem(4096, iobuf.PageSize)
	}
}

func BenchmarkAlignedMem_64K(b *testing.B) {
	for range b.N {
		_ = iobuf.AlignedMem(65536, iobuf.PageSize)
	}
}

func BenchmarkAlignedMemBlocks_16(b *testing.B) {
	for range b.N {
		_ = iobuf.AlignedMemBlocks(16, iobuf.PageSize)
	}
}

// IoVec benchmarks

func BenchmarkIoVecFrom_SmallBuffer_8(b *testing.B) {
	buffers := make([]iobuf.SmallBuffer, 8)
	b.ResetTimer()
	for range b.N {
		_ = iobuf.IoVecFrom(buffers)
	}
}

func BenchmarkIoVecFrom_SmallBuffer_64(b *testing.B) {
	buffers := make([]iobuf.SmallBuffer, 64)
	b.ResetTimer()
	for range b.N {
		_ = iobuf.IoVecFrom(buffers)
	}
}

func BenchmarkIoVecFrom_LargeBuffer_8(b *testing.B) {
	buffers := make([]iobuf.LargeBuffer, 8)
	b.ResetTimer()
	for range b.N {
		_ = iobuf.IoVecFrom(buffers)
	}
}

func BenchmarkIoVecFromBytesSlice_8(b *testing.B) {
	slices := make([][]byte, 8)
	for i := range slices {
		slices[i] = make([]byte, 256)
	}
	b.ResetTimer()
	for range b.N {
		_ = iobuf.IoVecFromBytesSlice(slices)
	}
}

func BenchmarkIoVecAddrLen(b *testing.B) {
	buffers := make([]iobuf.SmallBuffer, 8)
	iovecs := iobuf.IoVecFrom(buffers)
	b.ResetTimer()
	for range b.N {
		_, _ = iobuf.IoVecAddrLen(iovecs)
	}
}

// Buffer value access benchmarks

func BenchmarkPool_Value(b *testing.B) {
	pool := iobuf.NewSmallBufferPool(1024)
	pool.Fill(iobuf.NewSmallBuffer)

	b.ResetTimer()
	for i := range b.N {
		_ = pool.Value(i % 1024)
	}
}

func BenchmarkPool_SetValue(b *testing.B) {
	pool := iobuf.NewSmallBufferPool(1024)
	pool.Fill(iobuf.NewSmallBuffer)
	buf := iobuf.NewSmallBuffer()

	b.ResetTimer()
	for i := range b.N {
		pool.SetValue(i%1024, buf)
	}
}

// High-contention benchmarks
//
// These benchmarks measure pool performance under contention with small pools
// and high parallelism.

func BenchmarkPool_HighContention_SmallPool(b *testing.B) {
	// Small pool (16 buffers) with high parallelism creates contention
	pool := iobuf.NewSmallBufferPool(16)
	pool.Fill(iobuf.NewSmallBuffer)
	pool.SetNonblock(true)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				continue
			}
			_ = pool.Put(idx)
		}
	})
}

func BenchmarkPool_HighContention_TinyPool(b *testing.B) {
	// Tiny pool (4 buffers) creates extreme contention
	pool := iobuf.NewSmallBufferPool(4)
	pool.Fill(iobuf.NewSmallBuffer)
	pool.SetNonblock(true)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				continue
			}
			_ = pool.Put(idx)
		}
	})
}

func BenchmarkPool_Contention_MediumBuffer(b *testing.B) {
	// Medium buffers with moderate contention
	pool := iobuf.NewMediumBufferPool(32)
	pool.Fill(iobuf.NewMediumBuffer)
	pool.SetNonblock(true)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				continue
			}
			_ = pool.Put(idx)
		}
	})
}

func BenchmarkPool_Contention_LargeBuffer(b *testing.B) {
	// Large buffers with moderate contention
	pool := iobuf.NewLargeBufferPool(32)
	pool.Fill(iobuf.NewLargeBuffer)
	pool.SetNonblock(true)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				continue
			}
			_ = pool.Put(idx)
		}
	})
}

// Cache-line aligned memory benchmarks

func BenchmarkCacheLineAlignedMem_64(b *testing.B) {
	for range b.N {
		_ = iobuf.CacheLineAlignedMem(64)
	}
}

func BenchmarkCacheLineAlignedMem_4K(b *testing.B) {
	for range b.N {
		_ = iobuf.CacheLineAlignedMem(4096)
	}
}

func BenchmarkCacheLineAlignedMemBlocks_8(b *testing.B) {
	for range b.N {
		_ = iobuf.CacheLineAlignedMemBlocks(8, 64)
	}
}

// Buffer tier selection benchmarks

func BenchmarkTierBySize(b *testing.B) {
	sizes := []int{16, 64, 256, 1024, 4096, 16384, 65536, 262144, 1048576}
	b.ResetTimer()
	for i := range b.N {
		_ = iobuf.TierBySize(sizes[i%len(sizes)])
	}
}

func BenchmarkBufferSizeFor(b *testing.B) {
	sizes := []int{16, 64, 256, 1024, 4096, 16384, 65536, 262144, 1048576}
	b.ResetTimer()
	for i := range b.N {
		_ = iobuf.BufferSizeFor(sizes[i%len(sizes)])
	}
}

func BenchmarkBufferTier_Size(b *testing.B) {
	tiers := []iobuf.BufferTier{
		iobuf.TierPico, iobuf.TierNano, iobuf.TierMicro, iobuf.TierSmall,
		iobuf.TierMedium, iobuf.TierBig, iobuf.TierLarge, iobuf.TierGreat,
	}
	b.ResetTimer()
	for i := range b.N {
		_ = tiers[i%len(tiers)].Size()
	}
}

// Sequential (single-threaded) pool benchmarks for baseline comparison

func BenchmarkPool_Sequential_GetPut(b *testing.B) {
	pool := iobuf.NewSmallBufferPool(1024)
	pool.Fill(iobuf.NewSmallBuffer)

	b.ResetTimer()
	for range b.N {
		idx, _ := pool.Get()
		_ = pool.Put(idx)
	}
}

func BenchmarkPool_Sequential_GetPut_Nonblocking(b *testing.B) {
	pool := iobuf.NewSmallBufferPool(1024)
	pool.Fill(iobuf.NewSmallBuffer)
	pool.SetNonblock(true)

	b.ResetTimer()
	for range b.N {
		idx, err := pool.Get()
		if err != nil {
			continue
		}
		_ = pool.Put(idx)
	}
}

// Per-tier pool benchmarks

func BenchmarkPicoBufferPool_GetPut(b *testing.B) {
	pool := iobuf.NewPicoBufferPool(1024)
	pool.Fill(iobuf.NewPicoBuffer)
	pool.SetNonblock(true)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				continue
			}
			_ = pool.Put(idx)
		}
	})
}

func BenchmarkNanoBufferPool_GetPut(b *testing.B) {
	pool := iobuf.NewNanoBufferPool(1024)
	pool.Fill(iobuf.NewNanoBuffer)
	pool.SetNonblock(true)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				continue
			}
			_ = pool.Put(idx)
		}
	})
}

func BenchmarkMicroBufferPool_GetPut(b *testing.B) {
	pool := iobuf.NewMicroBufferPool(1024)
	pool.Fill(iobuf.NewMicroBuffer)
	pool.SetNonblock(true)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				continue
			}
			_ = pool.Put(idx)
		}
	})
}

func BenchmarkBigBufferPool_GetPut(b *testing.B) {
	pool := iobuf.NewBigBufferPool(256)
	pool.Fill(iobuf.NewBigBuffer)
	pool.SetNonblock(true)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				continue
			}
			_ = pool.Put(idx)
		}
	})
}

// Buffer construction benchmarks

func BenchmarkNewBuffers_8x1K(b *testing.B) {
	for range b.N {
		_ = iobuf.NewBuffers(8, 1024)
	}
}

func BenchmarkNewBuffers_64x4K(b *testing.B) {
	for range b.N {
		_ = iobuf.NewBuffers(64, 4096)
	}
}

// Array conversion benchmarks

func BenchmarkSmallArrayFromSlice(b *testing.B) {
	data := make([]byte, iobuf.BufferSizeSmall*2)
	b.ResetTimer()
	for range b.N {
		_ = iobuf.ArrayFromSlice[iobuf.SmallBuffer](data, 0)
	}
}

func BenchmarkSliceOfSmallArray(b *testing.B) {
	data := make([]byte, iobuf.BufferSizeSmall*16)
	b.ResetTimer()
	for range b.N {
		_ = iobuf.SliceOfArray[iobuf.SmallBuffer](data, 0, 8)
	}
}

// Nonblocking pool benchmarks (parallel)

func BenchmarkPool_Nonblocking_Parallel(b *testing.B) {
	pool := iobuf.NewSmallBufferPool(1024)
	pool.Fill(iobuf.NewSmallBuffer)
	pool.SetNonblock(true)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx, err := pool.Get()
			if err != nil {
				continue
			}
			_ = pool.Put(idx)
		}
	})
}

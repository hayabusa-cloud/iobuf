// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package iobuf provides lock-free buffer pools and memory management utilities
// for high-performance I/O operations.
//
// The package implements a 12-tier buffer size hierarchy and lock-free bounded
// pools optimized for zero-allocation hot paths. All pools use semantic error
// types from iox for non-blocking control flow.
//
// # Buffer Tiers
//
// Buffers are organized into 12 size tiers following a power-of-4 progression:
//
//	Tier      Size       Use Case
//	────      ────       ────────
//	Pico      32 B       Tiny metadata, flags
//	Nano      128 B      Small headers, control frames
//	Micro     512 B      Protocol frames, small messages
//	Small     2 KiB      Typical network packets
//	Medium    8 KiB      Stream buffers, large packets
//	Big       32 KiB     TLS records, stream chunks
//	Large     128 KiB    io_uring buffer rings
//	Great     512 KiB    Large transfers
//	Huge      2 MiB      Huge page aligned buffers
//	Vast      8 MiB      Large file chunks
//	Giant     32 MiB     Video frames, datasets
//	Titan     128 MiB    Maximum allocation tier
//
// Each tier has corresponding type aliases (e.g., SmallBuffer, LargeBuffer) and
// factory functions for bounded pools (e.g., NewSmallBufferPool).
//
// # Bounded Pool
//
// BoundedPool uses a multi-producer multi-consumer (MPMC) index queue based on
// "A Scalable, Portable, and Memory-Efficient Lock-Free FIFO Queue" (Ruslan
// Nikolaev, 2019). Capacity rounds up to a power of two. The queue stores
// indices separately from the items and updates entries with atomic
// compare-and-swap. Slot remapping spreads adjacent cursors across cache lines.
//
// # Indirect Pool Pattern
//
// Get and Put transfer slot indices. Value returns the item by value. For a
// fixed-size buffer such as SmallBuffer, use SetValue to store a modified copy.
//
//	pool := NewSmallBufferPool(100)
//	pool.Fill(NewSmallBuffer)
//	idx, _ := pool.Get() // The freshly filled pool has an available slot.
//	buf := pool.Value(idx)
//	buf[0] = 1
//	pool.SetValue(idx, buf)
//	pool.Put(idx)
//
// To work on pooled bytes in place, fill a BoundedPool[[]byte] with a distinct
// backing buffer for each slot. Return the index once all application and
// kernel access to its bytes has ended.
//
// # Page-Aligned Memory
//
// For DMA and io_uring operations requiring page alignment:
//
//	mem := AlignedMem(4096, PageSize)  // Returns page-aligned []byte
//	block := AlignedMemBlock()         // Single page using default PageSize
//	blocks := AlignedMemBlocks(16, PageSize) // Multiple aligned blocks
//
// # Vectored I/O
//
// IoVec provides scatter/gather I/O support for readv/writev syscalls:
//
//	buffers := make([]SmallBuffer, 8)
//	iovecs := IoVecFrom(buffers)
//	addr, n := IoVecAddrLen(iovecs)  // Get pointer for syscall
//
// # Architecture Requirements
//
// This package requires a 64-bit CPU architecture (amd64, arm64, riscv64, loong64,
// ppc64, ppc64le, s390x, mips64, mips64le). 32-bit architectures are not supported
// due to 64-bit atomic operations in BoundedPool.
//
// # Thread Safety
//
// Fill the pool and configure its mode before sharing it. Multiple goroutines
// can then call Get and Put concurrently. In nonblocking mode an empty Get
// or full Put returns iox.ErrWouldBlock. Either mode may retry a contended
// queue update.
//
// # Dependencies
//
// iox supplies ErrWouldBlock for nonblocking calls and Backoff for blocking
// waits on an empty or full pool. spin supplies Wait for index queue contention.
package iobuf

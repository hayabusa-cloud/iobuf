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
// BoundedPool is a lock-free multi-producer multi-consumer (MPMC) pool based on
// the algorithm from "A Scalable, Portable, and Memory-Efficient Lock-Free FIFO
// Queue" (Ruslan Nikolaev, 2019). Key characteristics:
//
//   - Lock-free: Uses atomic CAS operations, no mutexes
//   - Bounded: Fixed capacity rounded to power of two
//   - Memory-efficient: Single contiguous array, no per-element allocation
//   - Cache-optimized: Aligned to cache line boundaries to prevent false sharing
//
// # Indirect Pool Pattern
//
// Pools store indices (int) rather than buffer values directly. This enables:
//
//   - Zero-copy access via Value(indirect) method
//   - Efficient pool operations without moving large buffers
//   - Clear ownership semantics through index hand-off
//
// Usage pattern:
//
//	pool := NewSmallBufferPool(100) // Creates pool with ~128 capacity
//	pool.Fill(NewSmallBuffer)       // Initialize with buffer factory
//	idx, err := pool.Get()          // Acquire buffer index
//	if err != nil {
//	    // Handle iox.ErrWouldBlock (pool empty)
//	}
//	buf := pool.Value(idx)          // Access buffer by index
//	// Use buf[:]...
//	pool.Put(idx)                   // Return buffer to pool
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
//	iovecs := IoVecFromSmallBuffers(buffers)
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
// All pool operations are safe for concurrent use. BoundedPool supports multiple
// concurrent producers and consumers without external synchronization.
//
// # Dependencies
//
// iobuf depends on:
//   - iox: Semantic error types (ErrWouldBlock, ErrMore)
//   - spin: Spinlock and spin-wait primitives for backpressure
package iobuf

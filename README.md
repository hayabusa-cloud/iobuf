# iobuf

[![Go Reference](https://pkg.go.dev/badge/code.hybscloud.com/iobuf.svg)](https://pkg.go.dev/code.hybscloud.com/iobuf)
[![Go Report Card](https://goreportcard.com/badge/github.com/hayabusa-cloud/iobuf)](https://goreportcard.com/report/github.com/hayabusa-cloud/iobuf)
[![codecov](https://codecov.io/gh/hayabusa-cloud/iobuf/branch/main/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/iobuf)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Lock-free, memory-friendly bounded buffer pools for Go optimized for low-latency systems.

English | [简体中文](README.zh-CN.md) | [Español](README.es.md) | [日本語](README.ja.md) | [Français](README.fr.md)

## Three-Tier Progress Model

`iobuf` utilizes the **Spin** and **Adapt** layers of our performance ecosystem:

1.  **Strike**: System call → Direct kernel hit.
2.  **Spin**: Hardware yield (`spin`) → Local atomic synchronization.
3.  **Adapt**: Software backoff (`iox.Backoff`) → External I/O readiness.

## Features

- **Bounded lock-free buffer pools** for low-latency systems.
- **Page-aligned memory allocation** for DMA and io_uring compatibility.
- **Zero-copy IoVec generation** for vectored I/O syscalls.
- **Cooperative back-off**: Uses `iox.Backoff` to handle resource exhaustion gracefully.

## Requirements

- **Go 1.25+**
- **64-bit CPU** (amd64, arm64, riscv64, loong64, ppc64, s390x, mips64, etc.)

> **Note:** 32-bit architectures are not supported due to 64-bit atomic operations in the lock-free pool implementation.

## Installation

```bash
go get code.hybscloud.com/iobuf
```

## Quick Start

### Buffer Pools

```go
// Create a pool of 1024 small buffers (2 KiB each)
pool := iobuf.NewSmallBufferPool(1024)
pool.Fill(iobuf.NewSmallBuffer)

// Acquire a buffer index
idx, err := pool.Get()
if err != nil {
    panic(err)
}

// Access the buffer directly (zero-copy)
buf := pool.Value(idx)
...

// Return to pool
pool.Put(idx)
```

### Page-Aligned Memory

```go
// Single page-aligned block (default page size)
block := iobuf.AlignedMemBlock()

// Custom size with explicit alignment
mem := iobuf.AlignedMem(65536, iobuf.PageSize)

// Multiple aligned blocks
blocks := iobuf.AlignedMemBlocks(16, iobuf.PageSize)
```

### IoVec for Vectored I/O

```go
// Convert tiered buffers to iovec for readv/writev
buffers := make([]iobuf.SmallBuffer, 8)
iovecs := iobuf.IoVecFromSmallBuffers(buffers)

// Get raw pointer and count for syscalls
addr, n := iobuf.IoVecAddrLen(iovecs)
```

## Buffer Tiers

Power-of-4 progression starting at 32 bytes (12 tiers, 32 B to 128 MiB):

| Tier | Size | Use Case |
|------|------|----------|
| Pico | 32 B | UUIDs, flags, tiny control messages |
| Nano | 128 B | HTTP headers, JSON tokens, small RPC payloads |
| Micro | 512 B | DNS packets, MQTT messages, protocol frames |
| Small | 2 KiB | WebSocket frames, small HTTP responses |
| Medium | 8 KiB | TCP segments, gRPC messages, page I/O |
| Big | 32 KiB | TLS records (16 KiB max), stream chunks |
| Large | 128 KiB | io_uring buffer rings, bulk network transfers |
| Great | 512 KiB | Database pages, large API responses |
| Huge | 2 MiB | Huge page aligned, memory-mapped files |
| Vast | 8 MiB | Image processing, compressed archives |
| Giant | 32 MiB | Video frames, ML model weights |
| Titan | 128 MiB | Large datasets, maximum stack-safe buffer |

## API Overview

### Pool Interfaces

```go
// Generic pool interface
type Pool[T any] interface {
    Put(item T) error
    Get() (item T, err error)
}

// Index-based pool for zero-copy buffer access
type IndirectPool[T BufferType] interface {
    Pool[int]
    Value(indirect int) T
    SetValue(indirect int, item T)
}
```

### Pool Constructors

```go
func NewPicoBufferPool(capacity int) *PicoBufferBoundedPool
func NewNanoBufferPool(capacity int) *NanoBufferBoundedPool
func NewMicroBufferPool(capacity int) *MicroBufferBoundedPool
func NewSmallBufferPool(capacity int) *SmallBufferBoundedPool
func NewMediumBufferPool(capacity int) *MediumBufferBoundedPool
func NewBigBufferPool(capacity int) *BigBufferBoundedPool
func NewLargeBufferPool(capacity int) *LargeBufferBoundedPool
func NewGreatBufferPool(capacity int) *GreatBufferBoundedPool
func NewHugeBufferPool(capacity int) *HugeBufferBoundedPool
func NewVastBufferPool(capacity int) *VastBufferBoundedPool
func NewGiantBufferPool(capacity int) *GiantBufferBoundedPool
func NewTitanBufferPool(capacity int) *TitanBufferBoundedPool
```

### Memory Allocation

```go
// Page-aligned memory
func AlignedMem(size int, pageSize uintptr) []byte
func AlignedMemBlocks(n int, pageSize uintptr) [][]byte
func AlignedMemBlock() []byte

// Cache-line-aligned memory (prevents false sharing)
func CacheLineAlignedMem(size int) []byte
func CacheLineAlignedMemBlocks(n int, blockSize int) [][]byte
const CacheLineSize  // 64 or 128 depending on architecture
```

### IoVec Generation

```go
func IoVecFromBytesSlice(iov [][]byte) (addr uintptr, n int)
func IoVecFromPicoBuffers(buffers []PicoBuffer) []IoVec
// ... and for all other tiers
```

## Design

The bounded pool implementation is based on lock-free queue algorithms:

- **Memory-efficient**: O(n) space for n-capacity pool
- **Lock-free progress**: Guaranteed global progress bounds
- **Cache-friendly**: Minimizes false sharing and cache-line bouncing

## References

- [Morrison & Afek, "Fast concurrent queues for x86 processors," PPoPP 2013](https://dl.acm.org/doi/10.1145/2442516.2442527)
- [Nikolaev, "A scalable, portable, and memory-efficient lock-free FIFO queue," DISC 2019](https://drops.dagstuhl.de/opus/volltexte/2019/11335/pdf/LIPIcs-DISC-2019-28.pdf)
- [Koval & Aksenov, "Restricted memory-friendly lock-free bounded queues," PPoPP 2020](https://nikitakoval.org/publications/ppopp20-queues.pdf)
- [Nikolaev & Ravindran, "wCQ: A fast wait-free queue with bounded memory usage," 2022](https://arxiv.org/abs/2201.02179)
- [Aksenov et al., "Memory bounds for concurrent bounded queues," 2024](https://arxiv.org/abs/2104.15003)
- [Denis & Goedefroit, "NBLFQ: A lock-free MPMC queue optimized for low contention," IPDPS 2025](https://hal.science/hal-04762608)

## License

MIT License - see [LICENSE](LICENSE) for details.

© 2025 [Hayabusa Cloud Co., Ltd.](https://code.hybscloud.com)

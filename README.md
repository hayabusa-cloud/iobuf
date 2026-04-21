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

- **Go 1.26+**
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
iovecs := iobuf.IoVecFrom(buffers)

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
func IoVecFrom[T BufferType](buffers []T) []IoVec
func IoVecFromBytesSlice(iov [][]byte) (addr uintptr, n int)
func IoVecFromRegisteredBuffers(buffers []RegisterBuffer) []IoVec
func IoVecAddrLen(vec []IoVec) (addr uintptr, n int)
```

## Design

The bounded pool implementation is based on lock-free queue algorithms:

- **Memory-efficient**: O(n) space for n-capacity pool
- **Lock-free progress**: Guaranteed global progress bounds
- **Cache-friendly**: Minimizes false sharing and cache-line bouncing

## References

- Adam Morrison and Yehuda Afek. 2013. Fast Concurrent Queues for x86 Processors. In *Proc. 18th ACM SIGPLAN Symposium
  on Principles and Practice of Parallel Programming (PPoPP '13)*. 103–112. https://doi.org/10.1145/2442516.2442527
- Ruslan Nikolaev. 2019. A Scalable, Portable, and Memory-Efficient Lock-Free FIFO Queue. In *33rd International
  Symposium on Distributed Computing (DISC 2019)*. Leibniz International Proceedings in Informatics (LIPIcs) 146, 28:
  1–28:16. https://arxiv.org/abs/1908.04511
- Nikita Koval and Vitaly Aksenov. 2020. POSTER: Restricted Memory-Friendly Lock-Free Bounded Queues. In *Proceedings of
  the 25th ACM SIGPLAN Symposium on Principles and Practice of Parallel Programming (PPoPP '20), February 22–26, 2020,
  San Diego, CA, USA*. Association for Computing Machinery, New York, NY, USA, 433–434.
  https://doi.org/10.1145/3332466.3374508
- Ruslan Nikolaev and Binoy Ravindran. 2022. wCQ: A Fast Wait-Free Queue with Bounded Memory Usage. In *Proc. 34th ACM
  Symposium on Parallelism in Algorithms and Architectures (SPAA '22)*. 307–319. https://arxiv.org/abs/2201.02179
- Vitaly Aksenov, Nikita Koval, Petr Kuznetsov, and Anton Paramonov. 2024. Memory Bounds for Concurrent Bounded Queues.
  In *Proc. 29th ACM SIGPLAN Annual Symposium on Principles and Practice of Parallel Programming (PPoPP '24)*.
  188–199. https://arxiv.org/abs/2104.15003
- Alexandre Denis and Charles Goedefroit. 2025. NBLFQ: A Lock-Free MPMC Queue Optimized for Low Contention. In *2025
  IEEE International Parallel and Distributed Processing Symposium (IPDPS 2025)*.
  962–973. https://inria.hal.science/hal-04851700/file/article-final.pdf

## License

MIT License - see [LICENSE](LICENSE) for details.

© 2025 [Hayabusa Cloud Co., Ltd.](https://code.hybscloud.com)

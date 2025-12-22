# iobuf

[![Go Reference](https://pkg.go.dev/badge/code.hybscloud.com/iobuf.svg)](https://pkg.go.dev/code.hybscloud.com/iobuf)
[![Go Report Card](https://goreportcard.com/badge/github.com/hayabusa-cloud/iobuf)](https://goreportcard.com/report/github.com/hayabusa-cloud/iobuf)
[![codecov](https://codecov.io/gh/hayabusa-cloud/iobuf/branch/main/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/iobuf)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

面向低延迟系统优化的无锁、内存友好型有界缓冲池。

[English](README.md) | 简体中文 | [Español](README.es.md) | [日本語](README.ja.md) | [Français](README.fr.md)

## 三层进展模型

`iobuf` 使用我们性能生态系统中的 **Spin** 和 **Adapt** 层：

1.  **Strike**：系统调用 → 直接内核命中。
2.  **Spin**：硬件让步 (`spin`) → 本地原子同步。
3.  **Adapt**：软件退避 (`iox.Backoff`) → 外部 I/O 就绪。

## 特性

- **有界无锁缓冲池**：面向低延迟系统。
- **页对齐内存分配**：兼容 DMA 和 io_uring。
- **零拷贝 IoVec 生成**：用于向量化 I/O 系统调用。
- **协作式退避**：使用 `iox.Backoff` 优雅处理资源耗尽。

## 安装

```bash
go get code.hybscloud.com/iobuf
```

## 快速开始

### 缓冲池

```go
// 创建包含 1024 个小缓冲区的池（每个 1 KiB）
pool := iobuf.NewSmallBufferPool(1024)
pool.Fill(iobuf.NewSmallBuffer)

// 获取缓冲区索引
idx, err := pool.Get()
if err != nil {
    panic(err)
}

// 直接访问缓冲区（零拷贝）
buf := pool.Value(idx)
...

// 归还到池
pool.Put(idx)
```

### 页对齐内存

```go
// 单个页对齐块（默认页大小）
block := iobuf.AlignedMemBlock()

// 自定义大小和显式对齐
mem := iobuf.AlignedMem(65536, iobuf.PageSize)

// 多个对齐块
blocks := iobuf.AlignedMemBlocks(16, iobuf.PageSize)
```

### 用于向量化 I/O 的 IoVec

```go
// 将分层缓冲区转换为 iovec 用于 readv/writev
buffers := make([]iobuf.SmallBuffer, 8)
iovecs := iobuf.IoVecFromSmallBuffers(buffers)

// 获取原始指针和计数用于系统调用
addr, n := iobuf.IoVecAddrLen(iovecs)
```

## 缓冲区层级

4 的幂次递增，从 16 字节开始：

| 层级 | 大小 | 用途 |
|------|------|------|
| Pico | 16 B | 微型元数据、标志位 |
| Nano | 64 B | 小型头部、令牌 |
| Micro | 256 B | 协议头部 |
| Small | 1 KiB | 小型消息 |
| Medium | 4 KiB | 页大小 I/O |
| Large | 16 KiB | 大型传输 |
| Huge | 64 KiB | 最大 UDP |
| Giant | 256 KiB | 批量 I/O、大型负载 |

## API 概览

### 池接口

```go
// 通用池接口
type Pool[T any] interface {
    Put(item T) error
    Get() (item T, err error)
}

// 基于索引的池，用于零拷贝缓冲区访问
type IndirectPool[T BufferType] interface {
    Pool[int]
    Value(indirect int) T
    SetValue(indirect int, item T)
}
```

### 池构造函数

```go
func NewPicoBufferPool(capacity int) *PicoBufferBoundedPool
func NewNanoBufferPool(capacity int) *NanoBufferBoundedPool
func NewMicroBufferPool(capacity int) *MicroBufferBoundedPool
func NewSmallBufferPool(capacity int) *SmallBufferBoundedPool
func NewMediumBufferPool(capacity int) *MediumBufferBoundedPool
func NewLargeBufferPool(capacity int) *LargeBufferBoundedPool
func NewHugeBufferPool(capacity int) *HugeBufferBoundedPool
func NewGiantBufferPool(capacity int) *GiantBufferBoundedPool
```

### 内存分配

```go
func AlignedMem(size int, pageSize uintptr) []byte
func AlignedMemBlocks(n int, pageSize uintptr) [][]byte
func AlignedMemBlock() []byte
```

### IoVec 生成

```go
func IoVecFromBytesSlice(iov [][]byte) (addr uintptr, n int)
func IoVecFromPicoBuffers(buffers []PicoBuffer) []IoVec
// ... 以及所有其他层级
```

## 设计

有界池实现基于无锁队列算法：

- **内存高效**：n 容量池的 O(n) 空间复杂度
- **无锁进展**：保证全局进展边界
- **缓存友好**：最小化伪共享和缓存行抖动

## 参考文献

- [Morrison & Afek, "Fast concurrent queues for x86 processors," PPoPP 2013](https://dl.acm.org/doi/10.1145/2442516.2442527)
- [Nikolaev, "A scalable, portable, and memory-efficient lock-free FIFO queue," DISC 2019](https://drops.dagstuhl.de/opus/volltexte/2019/11335/pdf/LIPIcs-DISC-2019-28.pdf)
- [Koval & Aksenov, "Restricted memory-friendly lock-free bounded queues," PPoPP 2020](https://nikitakoval.org/publications/ppopp20-queues.pdf)
- [Nikolaev & Ravindran, "wCQ: A fast wait-free queue with bounded memory usage," 2022](https://arxiv.org/abs/2201.02179)
- [Aksenov et al., "Memory bounds for concurrent bounded queues," 2024](https://arxiv.org/abs/2104.15003)
- [Denis & Goedefroit, "NBLFQ: A lock-free MPMC queue optimized for low contention," IPDPS 2025](https://hal.science/hal-04762608)

## 许可证

MIT 许可证 - 详见 [LICENSE](LICENSE)。

© 2025 [Hayabusa Cloud Co., Ltd.](https://code.hybscloud.com)

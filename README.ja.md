# iobuf

[![Go Reference](https://pkg.go.dev/badge/code.hybscloud.com/iobuf.svg)](https://pkg.go.dev/code.hybscloud.com/iobuf)
[![Go Report Card](https://goreportcard.com/badge/github.com/hayabusa-cloud/iobuf)](https://goreportcard.com/report/github.com/hayabusa-cloud/iobuf)
[![codecov](https://codecov.io/gh/hayabusa-cloud/iobuf/branch/main/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/iobuf)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

低レイテンシシステム向けに最適化された、ロックフリーでメモリ効率の良い有界バッファプール。

[English](README.md) | [简体中文](README.zh-CN.md) | [Español](README.es.md) | 日本語 | [Français](README.fr.md)

## 三層進行モデル

`iobuf` はパフォーマンスエコシステムの **Spin** と **Adapt** レイヤーを使用します：

1.  **Strike**：システムコール → カーネルへの直接ヒット。
2.  **Spin**：ハードウェアイールド (`spin`) → ローカルアトミック同期。
3.  **Adapt**：ソフトウェアバックオフ (`iox.Backoff`) → 外部I/O準備完了。

## 特徴

- **有界ロックフリーバッファプール**：低レイテンシシステム向け。
- **ページアラインメモリ割り当て**：DMAおよびio_uring互換。
- **ゼロコピーIoVec生成**：ベクトル化I/Oシステムコール用。
- **協調的バックオフ**：`iox.Backoff` を使用してリソース枯渇を優雅に処理。

## インストール

```bash
go get code.hybscloud.com/iobuf
```

## クイックスタート

### バッファプール

```go
// 1024個のスモールバッファ（各1 KiB）のプールを作成
pool := iobuf.NewSmallBufferPool(1024)
pool.Fill(iobuf.NewSmallBuffer)

// バッファインデックスを取得
idx, err := pool.Get()
if err != nil {
    panic(err)
}

// バッファに直接アクセス（ゼロコピー）
buf := pool.Value(idx)
...

// プールに返却
pool.Put(idx)
```

### ページアラインメモリ

```go
// 単一のページアラインブロック（デフォルトページサイズ）
block := iobuf.AlignedMemBlock()

// カスタムサイズと明示的アラインメント
mem := iobuf.AlignedMem(65536, iobuf.PageSize)

// 複数のアラインブロック
blocks := iobuf.AlignedMemBlocks(16, iobuf.PageSize)
```

### ベクトル化I/O用IoVec

```go
// 階層化バッファをreadv/writev用のiovecに変換
buffers := make([]iobuf.SmallBuffer, 8)
iovecs := iobuf.IoVecFromSmallBuffers(buffers)

// システムコール用の生ポインタとカウントを取得
addr, n := iobuf.IoVecAddrLen(iovecs)
```

## バッファ階層

16バイトから始まる4の累乗で増加：

| 階層 | サイズ | 用途 |
|------|--------|------|
| Pico | 16 B | 小さなメタデータ、フラグ |
| Nano | 64 B | 小さなヘッダ、トークン |
| Micro | 256 B | プロトコルヘッダ |
| Small | 1 KiB | 小さなメッセージ |
| Medium | 4 KiB | ページサイズI/O |
| Large | 16 KiB | 大きな転送 |
| Huge | 64 KiB | 最大UDP |
| Giant | 256 KiB | バルクI/O、大きなペイロード |

## API概要

### プールインターフェース

```go
// 汎用プールインターフェース
type Pool[T any] interface {
    Put(item T) error
    Get() (item T, err error)
}

// ゼロコピーバッファアクセス用のインデックスベースプール
type IndirectPool[T BufferType] interface {
    Pool[int]
    Value(indirect int) T
    SetValue(indirect int, item T)
}
```

### プールコンストラクタ

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

### メモリ割り当て

```go
func AlignedMem(size int, pageSize uintptr) []byte
func AlignedMemBlocks(n int, pageSize uintptr) [][]byte
func AlignedMemBlock() []byte
```

### IoVec生成

```go
func IoVecFromBytesSlice(iov [][]byte) (addr uintptr, n int)
func IoVecFromPicoBuffers(buffers []PicoBuffer) []IoVec
// ... および他のすべての階層
```

## 設計

有界プールの実装はロックフリーキューアルゴリズムに基づいています：

- **メモリ効率**：容量nのプールに対してO(n)空間
- **ロックフリー進行**：グローバル進行境界を保証
- **キャッシュフレンドリー**：フォルスシェアリングとキャッシュラインバウンシングを最小化

## 参考文献

- [Morrison & Afek, "Fast concurrent queues for x86 processors," PPoPP 2013](https://dl.acm.org/doi/10.1145/2442516.2442527)
- [Nikolaev, "A scalable, portable, and memory-efficient lock-free FIFO queue," DISC 2019](https://drops.dagstuhl.de/opus/volltexte/2019/11335/pdf/LIPIcs-DISC-2019-28.pdf)
- [Koval & Aksenov, "Restricted memory-friendly lock-free bounded queues," PPoPP 2020](https://nikitakoval.org/publications/ppopp20-queues.pdf)
- [Nikolaev & Ravindran, "wCQ: A fast wait-free queue with bounded memory usage," 2022](https://arxiv.org/abs/2201.02179)
- [Aksenov et al., "Memory bounds for concurrent bounded queues," 2024](https://arxiv.org/abs/2104.15003)
- [Denis & Goedefroit, "NBLFQ: A lock-free MPMC queue optimized for low contention," IPDPS 2025](https://hal.science/hal-04762608)

## ライセンス

MITライセンス - 詳細は[LICENSE](LICENSE)を参照してください。

© 2025 [Hayabusa Cloud Co., Ltd.](https://code.hybscloud.com)

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

## システム要件

- **Go 1.26+**
- **64ビットCPU**（amd64、arm64、riscv64、loong64、ppc64、s390x、mips64など）

> **注意：** ロックフリープール実装で64ビットアトミック操作を使用しているため、32ビットアーキテクチャはサポートされていません。

## インストール

```bash
go get code.hybscloud.com/iobuf
```

## クイックスタート

### バッファプール

```go
// 1024個のスモールバッファ（各2 KiB）のプールを作成
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
iovecs := iobuf.IoVecFrom(buffers)

// システムコール用の生ポインタとカウントを取得
addr, n := iobuf.IoVecAddrLen(iovecs)
```

## バッファ階層

32バイトから始まる4の累乗で増加（12階層、32 B から 128 MiB）：

| 階層 | サイズ | 用途 |
|------|--------|------|
| Pico | 32 B | UUID、フラグ、小さな制御メッセージ |
| Nano | 128 B | HTTPヘッダ、JSONトークン、小さなRPCペイロード |
| Micro | 512 B | DNSパケット、MQTTメッセージ、プロトコルフレーム |
| Small | 2 KiB | WebSocketフレーム、小さなHTTPレスポンス |
| Medium | 8 KiB | TCPセグメント、gRPCメッセージ、ページI/O |
| Big | 32 KiB | TLSレコード（最大16 KiB）、ストリームチャンク |
| Large | 128 KiB | io_uringバッファリング、バルクネットワーク転送 |
| Great | 512 KiB | データベースページ、大規模APIレスポンス |
| Huge | 2 MiB | ヒュージページ整列、メモリマップファイル |
| Vast | 8 MiB | 画像処理、圧縮アーカイブ |
| Giant | 32 MiB | ビデオフレーム、MLモデル重み |
| Titan | 128 MiB | 大規模データセット、最大スタック安全バッファ |

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
func NewBigBufferPool(capacity int) *BigBufferBoundedPool
func NewLargeBufferPool(capacity int) *LargeBufferBoundedPool
func NewGreatBufferPool(capacity int) *GreatBufferBoundedPool
func NewHugeBufferPool(capacity int) *HugeBufferBoundedPool
func NewVastBufferPool(capacity int) *VastBufferBoundedPool
func NewGiantBufferPool(capacity int) *GiantBufferBoundedPool
func NewTitanBufferPool(capacity int) *TitanBufferBoundedPool
```

### メモリ割り当て

```go
func AlignedMem(size int, pageSize uintptr) []byte
func AlignedMemBlocks(n int, pageSize uintptr) [][]byte
func AlignedMemBlock() []byte
```

### IoVec生成

```go
func IoVecFrom[T BufferType](buffers []T) []IoVec
func IoVecFromBytesSlice(iov [][]byte) (addr uintptr, n int)
func IoVecFromRegisteredBuffers(buffers []RegisterBuffer) []IoVec
func IoVecAddrLen(vec []IoVec) (addr uintptr, n int)
```

## 設計

有界プールの実装はロックフリーキューアルゴリズムに基づいています：

- **メモリ効率**：容量nのプールに対してO(n)空間
- **ロックフリー進行**：グローバル進行境界を保証
- **キャッシュフレンドリー**：フォルスシェアリングとキャッシュラインバウンシングを最小化

## 参考文献

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

## ライセンス

MITライセンス - 詳細は[LICENSE](LICENSE)を参照してください。

© 2025 [Hayabusa Cloud Co., Ltd.](https://code.hybscloud.com)

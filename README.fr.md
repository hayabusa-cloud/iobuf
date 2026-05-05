# iobuf

[![Go Reference](https://pkg.go.dev/badge/code.hybscloud.com/iobuf.svg)](https://pkg.go.dev/code.hybscloud.com/iobuf)
[![Go Report Card](https://goreportcard.com/badge/github.com/hayabusa-cloud/iobuf)](https://goreportcard.com/report/github.com/hayabusa-cloud/iobuf)
[![codecov](https://codecov.io/gh/hayabusa-cloud/iobuf/branch/main/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/iobuf)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Pools de buffers bornés sans verrou et économes en mémoire pour Go, optimisés pour les systèmes à faible latence.

[English](README.md) | [简体中文](README.zh-CN.md) | [Español](README.es.md) | [日本語](README.ja.md) | Français

## Modèle de Progression à Trois Niveaux

`iobuf` utilise les couches **Spin** et **Adapt** de notre écosystème de performance :

1.  **Strike** : Appel système → Impact direct au noyau.
2.  **Spin** : Cession matérielle (`spin`) → Synchronisation atomique locale.
3.  **Adapt** : Recul logiciel (`iox.Backoff`) → Préparation I/O externe.

## Caractéristiques

- **Pools de buffers bornés sans verrou** pour les systèmes à faible latence.
- **Allocation mémoire alignée sur page** compatible DMA et io_uring.
- **Génération IoVec sans copie** pour les appels système d'I/O vectorisées.
- **Recul coopératif** : Utilise `iox.Backoff` pour gérer l'épuisement des ressources avec élégance.

## Prérequis

- **Go 1.26+**
- **CPU 64 bits** (amd64, arm64, riscv64, loong64, ppc64, s390x, mips64, etc.)

> **Note :** Les architectures 32 bits ne sont pas prises en charge en raison des opérations atomiques 64 bits dans l'implémentation du pool sans verrou.

## Installation

```bash
go get code.hybscloud.com/iobuf
```

## Démarrage Rapide

### Pools de Buffers

```go
// Créer un pool de 1024 petits buffers (2 Kio chacun)
pool := iobuf.NewSmallBufferPool(1024)
pool.Fill(iobuf.NewSmallBuffer)

// Acquérir un index de buffer
idx, err := pool.Get()
if err != nil {
    panic(err)
}

// Accéder au buffer directement (sans copie)
buf := pool.Value(idx)
...

// Retourner au pool
pool.Put(idx)
```

### Mémoire Alignée sur Page

```go
// Bloc unique aligné sur page (taille de page par défaut)
block := iobuf.AlignedMemBlock()

// Taille personnalisée avec alignement explicite
mem := iobuf.AlignedMem(65536, iobuf.PageSize)

// Blocs multiples alignés
blocks := iobuf.AlignedMemBlocks(16, iobuf.PageSize)
```

### IoVec pour I/O Vectorisées

```go
// Convertir les buffers échelonnés en iovec pour readv/writev
buffers := make([]iobuf.SmallBuffer, 8)
iovecs := iobuf.IoVecFrom(buffers)

// Obtenir le pointeur brut et le compte pour les appels système
addr, n := iobuf.IoVecAddrLen(iovecs)
```

## Niveaux de Buffer

Progression en puissances de 4, à partir de 32 octets (12 niveaux, 32 o à 128 Mio) :

| Niveau | Taille | Cas d'Usage |
|--------|--------|-------------|
| Pico | 32 o | UUIDs, drapeaux, petits messages de contrôle |
| Nano | 128 o | En-têtes HTTP, jetons JSON, petits payloads RPC |
| Micro | 512 o | Paquets DNS, messages MQTT, trames de protocole |
| Small | 2 Kio | Frames WebSocket, petites réponses HTTP |
| Medium | 8 Kio | Segments TCP, messages gRPC, I/O de page |
| Big | 32 Kio | Enregistrements TLS (max 16 Kio), chunks de flux |
| Large | 128 Kio | Anneaux de tampon io_uring, transferts réseau massifs |
| Great | 512 Kio | Pages de base de données, grandes réponses API |
| Huge | 2 Mio | Aligné sur huge pages, fichiers mappés en mémoire |
| Vast | 8 Mio | Traitement d'images, archives compressées |
| Giant | 32 Mio | Frames vidéo, poids de modèles ML |
| Titan | 128 Mio | Grands ensembles de données, buffer max sûr pour pile |

## Aperçu de l'API

### Interfaces de Pool

```go
// Interface de pool générique
type Pool[T any] interface {
    Put(item T) error
    Get() (item T, err error)
}

// Pool basé sur index pour accès aux buffers sans copie
type IndirectPool[T BufferType] interface {
    Pool[int]
    Value(indirect int) T
    SetValue(indirect int, item T)
}
```

### Constructeurs de Pool

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

### Allocation Mémoire

```go
func AlignedMem(size int, pageSize uintptr) []byte
func AlignedMemBlocks(n int, pageSize uintptr) [][]byte
func AlignedMemBlock() []byte
```

### Génération IoVec

```go
func IoVecFrom[T BufferType](buffers []T) []IoVec
func IoVecFromBytesSlice(iov [][]byte) []IoVec
func IoVecFromRegisteredBuffers(buffers []RegisterBuffer) []IoVec
func IoVecAddrLen(vec []IoVec) (addr uintptr, n int)
```

## Conception

L'implémentation du pool borné est basée sur des algorithmes de files sans verrou :

- **Efficace en mémoire** : Espace O(n) pour un pool de capacité n
- **Progression sans verrou** : Bornes de progression globale garanties
- **Compatible cache** : Minimise le faux partage et le rebond de lignes de cache

## Références

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

## Licence

Licence MIT - voir [LICENSE](LICENSE) pour plus de détails.

© 2025 [Hayabusa Cloud Co., Ltd.](https://code.hybscloud.com)

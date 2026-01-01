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

- **Go 1.25+**
- **CPU 64 bits** (amd64, arm64, riscv64, loong64, ppc64, s390x, mips64, etc.)

> **Note :** Les architectures 32 bits ne sont pas prises en charge en raison des opérations atomiques 64 bits dans l'implémentation du pool sans verrou.

## Installation

```bash
go get code.hybscloud.com/iobuf
```

## Démarrage Rapide

### Pools de Buffers

```go
// Créer un pool de 1024 petits buffers (1 Kio chacun)
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
iovecs := iobuf.IoVecFromSmallBuffers(buffers)

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
func IoVecFromBytesSlice(iov [][]byte) (addr uintptr, n int)
func IoVecFromPicoBuffers(buffers []PicoBuffer) []IoVec
// ... et pour tous les autres niveaux
```

## Conception

L'implémentation du pool borné est basée sur des algorithmes de files sans verrou :

- **Efficace en mémoire** : Espace O(n) pour un pool de capacité n
- **Progression sans verrou** : Bornes de progression globale garanties
- **Compatible cache** : Minimise le faux partage et le rebond de lignes de cache

## Références

- [Morrison & Afek, "Fast concurrent queues for x86 processors," PPoPP 2013](https://dl.acm.org/doi/10.1145/2442516.2442527)
- [Nikolaev, "A scalable, portable, and memory-efficient lock-free FIFO queue," DISC 2019](https://drops.dagstuhl.de/opus/volltexte/2019/11335/pdf/LIPIcs-DISC-2019-28.pdf)
- [Koval & Aksenov, "Restricted memory-friendly lock-free bounded queues," PPoPP 2020](https://nikitakoval.org/publications/ppopp20-queues.pdf)
- [Nikolaev & Ravindran, "wCQ: A fast wait-free queue with bounded memory usage," 2022](https://arxiv.org/abs/2201.02179)
- [Aksenov et al., "Memory bounds for concurrent bounded queues," 2024](https://arxiv.org/abs/2104.15003)
- [Denis & Goedefroit, "NBLFQ: A lock-free MPMC queue optimized for low contention," IPDPS 2025](https://hal.science/hal-04762608)

## Licence

Licence MIT - voir [LICENSE](LICENSE) pour plus de détails.

© 2025 [Hayabusa Cloud Co., Ltd.](https://code.hybscloud.com)

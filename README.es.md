# iobuf

[![Go Reference](https://pkg.go.dev/badge/code.hybscloud.com/iobuf.svg)](https://pkg.go.dev/code.hybscloud.com/iobuf)
[![Go Report Card](https://goreportcard.com/badge/github.com/hayabusa-cloud/iobuf)](https://goreportcard.com/report/github.com/hayabusa-cloud/iobuf)
[![codecov](https://codecov.io/gh/hayabusa-cloud/iobuf/branch/main/graph/badge.svg)](https://codecov.io/gh/hayabusa-cloud/iobuf)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Pools de buffers acotados sin bloqueos y eficientes en memoria para Go, optimizados para sistemas de baja latencia.

[English](README.md) | [简体中文](README.zh-CN.md) | Español | [日本語](README.ja.md) | [Français](README.fr.md)

## Modelo de Progreso de Tres Niveles

`iobuf` utiliza las capas **Spin** y **Adapt** de nuestro ecosistema de rendimiento:

1.  **Strike**: Llamada al sistema → Impacto directo al kernel.
2.  **Spin**: Cesión de hardware (`spin`) → Sincronización atómica local.
3.  **Adapt**: Retroceso de software (`iox.Backoff`) → Preparación de I/O externa.

## Características

- **Pools de buffers acotados sin bloqueos** para sistemas de baja latencia.
- **Asignación de memoria alineada a página** compatible con DMA e io_uring.
- **Generación de IoVec sin copia** para llamadas al sistema de I/O vectorizado.
- **Retroceso cooperativo**: Usa `iox.Backoff` para manejar el agotamiento de recursos con elegancia.

## Requisitos

- **Go 1.26+**
- **CPU de 64 bits** (amd64, arm64, riscv64, loong64, ppc64, s390x, mips64, etc.)

> **Nota:** Las arquitecturas de 32 bits no son compatibles debido a las operaciones atómicas de 64 bits en la implementación del pool sin bloqueos.

## Instalación

```bash
go get code.hybscloud.com/iobuf
```

## Inicio Rápido

### Pools de Buffers

```go
// Crear un pool de 1024 buffers pequeños (2 KiB cada uno)
pool := iobuf.NewSmallBufferPool(1024)
pool.Fill(iobuf.NewSmallBuffer)

// Adquirir un índice de buffer
idx, err := pool.Get()
if err != nil {
    panic(err)
}

// Acceder al buffer directamente (sin copia)
buf := pool.Value(idx)
...

// Devolver al pool
pool.Put(idx)
```

### Memoria Alineada a Página

```go
// Bloque único alineado a página (tamaño de página predeterminado)
block := iobuf.AlignedMemBlock()

// Tamaño personalizado con alineación explícita
mem := iobuf.AlignedMem(65536, iobuf.PageSize)

// Múltiples bloques alineados
blocks := iobuf.AlignedMemBlocks(16, iobuf.PageSize)
```

### IoVec para I/O Vectorizado

```go
// Convertir buffers escalonados a iovec para readv/writev
buffers := make([]iobuf.SmallBuffer, 8)
iovecs := iobuf.IoVecFrom(buffers)

// Obtener puntero y contador para llamadas al sistema
addr, n := iobuf.IoVecAddrLen(iovecs)
```

## Niveles de Buffer

Progresión de potencias de 4, comenzando en 32 bytes (12 niveles, 32 B a 128 MiB):

| Nivel | Tamaño | Caso de Uso |
|-------|--------|-------------|
| Pico | 32 B | UUIDs, flags, mensajes de control pequeños |
| Nano | 128 B | Cabeceras HTTP, tokens JSON, payloads RPC pequeños |
| Micro | 512 B | Paquetes DNS, mensajes MQTT, tramas de protocolo |
| Small | 2 KiB | Frames WebSocket, respuestas HTTP pequeñas |
| Medium | 8 KiB | Segmentos TCP, mensajes gRPC, I/O de página |
| Big | 32 KiB | Registros TLS (máx 16 KiB), chunks de stream |
| Large | 128 KiB | Buffer rings io_uring, transferencias de red masivas |
| Great | 512 KiB | Páginas de base de datos, respuestas API grandes |
| Huge | 2 MiB | Alineado a huge pages, archivos mapeados en memoria |
| Vast | 8 MiB | Procesamiento de imágenes, archivos comprimidos |
| Giant | 32 MiB | Frames de video, pesos de modelos ML |
| Titan | 128 MiB | Datasets grandes, buffer máximo seguro para stack |

## Resumen de API

### Interfaces de Pool

```go
// Interfaz de pool genérica
type Pool[T any] interface {
    Put(item T) error
    Get() (item T, err error)
}

// Pool basado en índices para acceso a buffers sin copia
type IndirectPool[T BufferType] interface {
    Pool[int]
    Value(indirect int) T
    SetValue(indirect int, item T)
}
```

### Constructores de Pool

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

### Asignación de Memoria

```go
func AlignedMem(size int, pageSize uintptr) []byte
func AlignedMemBlocks(n int, pageSize uintptr) [][]byte
func AlignedMemBlock() []byte
```

### Generación de IoVec

```go
func IoVecFrom[T BufferType](buffers []T) []IoVec
func IoVecFromBytesSlice(iov [][]byte) []IoVec
func IoVecFromRegisteredBuffers(buffers []RegisterBuffer) []IoVec
func IoVecAddrLen(vec []IoVec) (addr uintptr, n int)
```

## Diseño

La implementación del pool acotado se basa en algoritmos de colas sin bloqueos:

- **Eficiente en memoria**: Espacio O(n) para pool de capacidad n
- **Progreso sin bloqueos**: Límites de progreso global garantizados
- **Amigable con caché**: Minimiza el false sharing y el rebote de líneas de caché

## Referencias

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

## Licencia

Licencia MIT - ver [LICENSE](LICENSE) para más detalles.

© 2025 [Hayabusa Cloud Co., Ltd.](https://code.hybscloud.com)

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

## Instalación

```bash
go get code.hybscloud.com/iobuf
```

## Inicio Rápido

### Pools de Buffers

```go
// Crear un pool de 1024 buffers pequeños (1 KiB cada uno)
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
iovecs := iobuf.IoVecFromSmallBuffers(buffers)

// Obtener puntero y contador para llamadas al sistema
addr, n := iobuf.IoVecAddrLen(iovecs)
```

## Niveles de Buffer

Progresión de potencias de 4, comenzando en 16 bytes:

| Nivel | Tamaño | Caso de Uso |
|-------|--------|-------------|
| Pico | 16 B | Metadatos pequeños, flags |
| Nano | 64 B | Cabeceras pequeñas, tokens |
| Micro | 256 B | Cabeceras de protocolo |
| Small | 1 KiB | Mensajes pequeños |
| Medium | 4 KiB | I/O de tamaño de página |
| Large | 16 KiB | Transferencias grandes |
| Huge | 64 KiB | UDP máximo |
| Giant | 256 KiB | I/O masivo, cargas grandes |

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
func NewLargeBufferPool(capacity int) *LargeBufferBoundedPool
func NewHugeBufferPool(capacity int) *HugeBufferBoundedPool
func NewGiantBufferPool(capacity int) *GiantBufferBoundedPool
```

### Asignación de Memoria

```go
func AlignedMem(size int, pageSize uintptr) []byte
func AlignedMemBlocks(n int, pageSize uintptr) [][]byte
func AlignedMemBlock() []byte
```

### Generación de IoVec

```go
func IoVecFromBytesSlice(iov [][]byte) (addr uintptr, n int)
func IoVecFromPicoBuffers(buffers []PicoBuffer) []IoVec
// ... y para todos los demás niveles
```

## Diseño

La implementación del pool acotado se basa en algoritmos de colas sin bloqueos:

- **Eficiente en memoria**: Espacio O(n) para pool de capacidad n
- **Progreso sin bloqueos**: Límites de progreso global garantizados
- **Amigable con caché**: Minimiza el false sharing y el rebote de líneas de caché

## Referencias

- [Morrison & Afek, "Fast concurrent queues for x86 processors," PPoPP 2013](https://dl.acm.org/doi/10.1145/2442516.2442527)
- [Nikolaev, "A scalable, portable, and memory-efficient lock-free FIFO queue," DISC 2019](https://drops.dagstuhl.de/opus/volltexte/2019/11335/pdf/LIPIcs-DISC-2019-28.pdf)
- [Koval & Aksenov, "Restricted memory-friendly lock-free bounded queues," PPoPP 2020](https://nikitakoval.org/publications/ppopp20-queues.pdf)
- [Nikolaev & Ravindran, "wCQ: A fast wait-free queue with bounded memory usage," 2022](https://arxiv.org/abs/2201.02179)
- [Aksenov et al., "Memory bounds for concurrent bounded queues," 2024](https://arxiv.org/abs/2104.15003)
- [Denis & Goedefroit, "NBLFQ: A lock-free MPMC queue optimized for low contention," IPDPS 2025](https://hal.science/hal-04762608)

## Licencia

Licencia MIT - ver [LICENSE](LICENSE) para más detalles.

© 2025 [Hayabusa Cloud Co., Ltd.](https://code.hybscloud.com)

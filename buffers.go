// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package iobuf

import (
	"unsafe"

	"code.hybscloud.com/iobuf/internal"
)

const (
	registerBufferSize = BufferSizeLarge
)

type (

	// RegisterBuffer represents a fixed-size buffer used for registering with the I/O ring.
	RegisterBuffer [registerBufferSize]byte

	// RegisterBufferPool represents a pool of fixed-size buffers used for registering with the I/O ring.
	RegisterBufferPool = BoundedPool[RegisterBuffer]
)

// AlignedMem returns a byte slice with the specified size
// and starting address aligned to the memory page size.
//
// This is useful for DMA operations and io_uring registered buffers
// that require page-aligned memory addresses.
//
// The returned slice shares underlying memory with a larger allocation;
// do not assume len(result) == cap(result).
func AlignedMem(size int, pageSize uintptr) []byte {
	p := make([]byte, uintptr(size)+pageSize-1)
	base := unsafe.Pointer(unsafe.SliceData(p))
	offset := ((uintptr(base)+pageSize-1)/pageSize)*pageSize - uintptr(base)
	return unsafe.Slice((*byte)(unsafe.Add(base, offset)), size)
}

// AlignedMemBlocks returns n page-aligned byte slices, each of length pageSize.
//
// All returned slices share a single contiguous underlying allocation,
// which is more memory-efficient than calling AlignedMem n times.
//
// Panics if n < 1.
func AlignedMemBlocks(n int, pageSize uintptr) (blocks [][]byte) {
	if n < 1 {
		panic("bad block num")
	}
	blocks = make([][]byte, n)
	p := make([]byte, int(pageSize)*(n+1))
	base := unsafe.Pointer(unsafe.SliceData(p))
	offset := ((uintptr(base)+pageSize-1)/pageSize)*pageSize - uintptr(base)
	for i := range n {
		blocks[i] = unsafe.Slice((*byte)(unsafe.Add(base, offset+uintptr(i)*pageSize)), pageSize)
	}
	return
}

// AlignedMemBlock returns a single page-aligned block using the system page size.
//
// This is a convenience function equivalent to AlignedMemBlocks(1, PageSize)[0].
func AlignedMemBlock() []byte {
	return AlignedMemBlocks(1, PageSize)[0]
}

// CacheLineSize is the CPU L1 cache line size for the current architecture.
// This is detected at compile time based on the target architecture:
//   - amd64: 64 bytes (Intel/AMD)
//   - arm64: 128 bytes (conservative for Apple Silicon)
//   - riscv64: 64 bytes
//   - loong64: 64 bytes
//   - others: 64 bytes (default)
const CacheLineSize = internal.CacheLineSize

// CacheLineAlignedMem returns a byte slice with the specified size
// and starting address aligned to the CPU cache line size.
// This is useful for preventing false sharing in concurrent data structures.
func CacheLineAlignedMem(size int) []byte {
	align := uintptr(CacheLineSize)
	p := make([]byte, uintptr(size)+align-1)
	base := unsafe.Pointer(unsafe.SliceData(p))
	offset := ((uintptr(base)+align-1)/align)*align - uintptr(base)
	return unsafe.Slice((*byte)(unsafe.Add(base, offset)), size)
}

// CacheLineAlignedMemBlocks returns n cache-line-aligned byte slices,
// each of length blockSize. Adjacent blocks are separated by cache line
// boundaries to prevent false sharing.
func CacheLineAlignedMemBlocks(n int, blockSize int) (blocks [][]byte) {
	if n < 1 {
		panic("bad block num")
	}
	align := uintptr(CacheLineSize)
	// Round up block size to cache line boundary
	alignedBlockSize := ((uintptr(blockSize) + align - 1) / align) * align
	totalSize := int(alignedBlockSize)*n + int(align) - 1
	p := make([]byte, totalSize)
	base := unsafe.Pointer(unsafe.SliceData(p))
	offset := ((uintptr(base)+align-1)/align)*align - uintptr(base)
	blocks = make([][]byte, n)
	for i := range n {
		blocks[i] = unsafe.Slice((*byte)(unsafe.Add(base, offset+uintptr(i)*alignedBlockSize)), blockSize)
	}
	return
}

// NewBuffers creates a Buffers slice containing n byte slices, each of length size.
//
// Returns an empty Buffers if n < 1. Each inner slice is independently allocated;
// for contiguous memory, use AlignedMemBlocks instead.
func NewBuffers(n int, size int) Buffers {
	if n < 1 {
		return Buffers{}
	}
	ret := make(Buffers, n)
	for i := range n {
		if size > 0 {
			ret[i] = make([]byte, size)
		} else {
			ret[i] = []byte{}
		}
	}

	return ret
}

// Buffer size tiers follow a power-of-4 progression starting at 32 bytes.
// Each tier is 4x the previous size, optimized for different I/O patterns.
// 12 tiers: 32B, 128B, 512B, 2KiB, 8KiB, 32KiB, 128KiB, 512KiB, 2MiB, 8MiB, 32MiB, 128MiB
const (
	BufferSizePico   = 1 << 5  // 32 B - tiny metadata, flags
	BufferSizeNano   = 1 << 7  // 128 B - small structs, headers
	BufferSizeMicro  = 1 << 9  // 512 B - protocol frames
	BufferSizeSmall  = 1 << 11 // 2 KiB - small messages
	BufferSizeMedium = 1 << 13 // 8 KiB - stream buffers
	BufferSizeBig    = 1 << 15 // 32 KiB - TLS records
	BufferSizeLarge  = 1 << 17 // 128 KiB - io_uring buffers
	BufferSizeGreat  = 1 << 19 // 512 KiB - large transfers
	BufferSizeHuge   = 1 << 21 // 2 MiB - huge pages
	BufferSizeVast   = 1 << 23 // 8 MiB - large file chunks
	BufferSizeGiant  = 1 << 25 // 32 MiB - video frames
	BufferSizeTitan  = 1 << 27 // 128 MiB - maximum buffer tier
)

// BufferTier represents a buffer tier index in the 12-tier system.
type BufferTier int

// Buffer tier indices for the 12-tier buffer system.
const (
	TierPico BufferTier = iota
	TierNano
	TierMicro
	TierSmall
	TierMedium
	TierBig
	TierLarge
	TierGreat
	TierHuge
	TierVast
	TierGiant
	TierTitan
	TierEnd // Sentinel marking end of tiers
)

// bufferSizes maps tier index to buffer size.
var bufferSizes = [TierEnd]int{
	TierPico:   BufferSizePico,
	TierNano:   BufferSizeNano,
	TierMicro:  BufferSizeMicro,
	TierSmall:  BufferSizeSmall,
	TierMedium: BufferSizeMedium,
	TierBig:    BufferSizeBig,
	TierLarge:  BufferSizeLarge,
	TierGreat:  BufferSizeGreat,
	TierHuge:   BufferSizeHuge,
	TierVast:   BufferSizeVast,
	TierGiant:  BufferSizeGiant,
	TierTitan:  BufferSizeTitan,
}

// TierBySize returns the smallest buffer tier that can hold 'size' bytes.
// Returns TierTitan for sizes larger than BufferSizeTitan.
func TierBySize(size int) BufferTier {
	switch {
	case size <= BufferSizePico:
		return TierPico
	case size <= BufferSizeNano:
		return TierNano
	case size <= BufferSizeMicro:
		return TierMicro
	case size <= BufferSizeSmall:
		return TierSmall
	case size <= BufferSizeMedium:
		return TierMedium
	case size <= BufferSizeBig:
		return TierBig
	case size <= BufferSizeLarge:
		return TierLarge
	case size <= BufferSizeGreat:
		return TierGreat
	case size <= BufferSizeHuge:
		return TierHuge
	case size <= BufferSizeVast:
		return TierVast
	case size <= BufferSizeGiant:
		return TierGiant
	default:
		return TierTitan
	}
}

// Size returns the buffer size for this tier.
func (t BufferTier) Size() int {
	if t < 0 || t >= TierEnd {
		return BufferSizeTitan
	}
	return bufferSizes[t]
}

// BufferSizeFor returns the smallest buffer size that can hold 'size' bytes.
// This is a convenience function equivalent to TierBySize(size).Size().
func BufferSizeFor(size int) int {
	return TierBySize(size).Size()
}

// NewPicoBuffer returns a zero-initialized PicoBuffer.
func NewPicoBuffer() PicoBuffer { return PicoBuffer{} }

// NewNanoBuffer returns a zero-initialized NanoBuffer.
func NewNanoBuffer() NanoBuffer { return NanoBuffer{} }

// NewMicroBuffer returns a zero-initialized MicroBuffer.
func NewMicroBuffer() MicroBuffer { return MicroBuffer{} }

// NewSmallBuffer returns a zero-initialized SmallBuffer.
func NewSmallBuffer() SmallBuffer { return SmallBuffer{} }

// NewMediumBuffer returns a zero-initialized MediumBuffer.
func NewMediumBuffer() MediumBuffer { return MediumBuffer{} }

// NewBigBuffer returns a zero-initialized BigBuffer.
func NewBigBuffer() BigBuffer { return BigBuffer{} }

// NewLargeBuffer returns a zero-initialized LargeBuffer.
func NewLargeBuffer() LargeBuffer { return LargeBuffer{} }

// NewGreatBuffer returns a zero-initialized GreatBuffer.
func NewGreatBuffer() GreatBuffer { return GreatBuffer{} }

// NewHugeBuffer returns a zero-initialized HugeBuffer.
func NewHugeBuffer() HugeBuffer { return HugeBuffer{} }

// NewVastBuffer returns a zero-initialized VastBuffer.
func NewVastBuffer() VastBuffer { return VastBuffer{} }

// NewGiantBuffer returns a zero-initialized GiantBuffer.
func NewGiantBuffer() GiantBuffer { return GiantBuffer{} }

// NewTitanBuffer returns a zero-initialized TitanBuffer.
func NewTitanBuffer() TitanBuffer { return TitanBuffer{} }

// BufferType is a type constraint for tiered buffer types.
type BufferType interface {
	PicoBuffer | NanoBuffer | MicroBuffer | SmallBuffer | MediumBuffer |
		BigBuffer | LargeBuffer | GreatBuffer | HugeBuffer | VastBuffer |
		GiantBuffer | TitanBuffer
}

type (
	// PicoBuffer is a 32-byte buffer for tiny metadata and flags.
	PicoBuffer [BufferSizePico]byte

	// NanoBuffer is a 128-byte buffer for small structs and headers.
	NanoBuffer [BufferSizeNano]byte

	// MicroBuffer is a 512-byte buffer for protocol frames.
	MicroBuffer [BufferSizeMicro]byte

	// SmallBuffer is a 2 KiB buffer for small messages.
	SmallBuffer [BufferSizeSmall]byte

	// MediumBuffer is an 8 KiB buffer for stream buffers.
	MediumBuffer [BufferSizeMedium]byte

	// BigBuffer is a 32 KiB buffer for TLS records.
	BigBuffer [BufferSizeBig]byte

	// LargeBuffer is a 128 KiB buffer for io_uring buffer rings.
	LargeBuffer [BufferSizeLarge]byte

	// GreatBuffer is a 512 KiB buffer for large transfers.
	GreatBuffer [BufferSizeGreat]byte

	// HugeBuffer is a 2 MiB buffer matching huge page sizes.
	HugeBuffer [BufferSizeHuge]byte

	// VastBuffer is an 8 MiB buffer for large file chunks.
	VastBuffer [BufferSizeVast]byte

	// GiantBuffer is a 32 MiB buffer for video frames and datasets.
	GiantBuffer [BufferSizeGiant]byte

	// TitanBuffer is a 128 MiB buffer, the maximum buffer tier.
	TitanBuffer [BufferSizeTitan]byte
)

// Reset methods are no-op implementations satisfying the Pool item contract.
// Buffer contents are not zeroed; callers should clear sensitive data explicitly.

func (b PicoBuffer) Reset()   {}
func (b NanoBuffer) Reset()   {}
func (b MicroBuffer) Reset()  {}
func (b SmallBuffer) Reset()  {}
func (b MediumBuffer) Reset() {}
func (b BigBuffer) Reset()    {}
func (b LargeBuffer) Reset()  {}
func (b GreatBuffer) Reset()  {}
func (b HugeBuffer) Reset()   {}
func (b VastBuffer) Reset()   {}
func (b GiantBuffer) Reset()  {}
func (b TitanBuffer) Reset()  {}

// PicoArrayFromSlice returns a PicoBuffer by copying from the slice at the given offset.
//
// The caller must ensure offset+BufferSizePico <= len(s).
// The returned array is a copy, not a view of the underlying slice.
func PicoArrayFromSlice(s []byte, offset int64) PicoBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizePico]byte)(ptr)
}

// NanoArrayFromSlice returns a NanoBuffer by copying from the slice at the given offset.
func NanoArrayFromSlice(s []byte, offset int64) NanoBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeNano]byte)(ptr)
}

// MicroArrayFromSlice returns a MicroBuffer by copying from the slice at the given offset.
func MicroArrayFromSlice(s []byte, offset int64) MicroBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeMicro]byte)(ptr)
}

// SmallArrayFromSlice returns a SmallBuffer by copying from the slice at the given offset.
func SmallArrayFromSlice(s []byte, offset int64) SmallBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeSmall]byte)(ptr)
}

// MediumArrayFromSlice returns a MediumBuffer by copying from the slice at the given offset.
func MediumArrayFromSlice(s []byte, offset int64) MediumBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeMedium]byte)(ptr)
}

// BigArrayFromSlice returns a BigBuffer by copying from the slice at the given offset.
func BigArrayFromSlice(s []byte, offset int64) BigBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeBig]byte)(ptr)
}

// LargeArrayFromSlice returns a LargeBuffer by copying from the slice at the given offset.
func LargeArrayFromSlice(s []byte, offset int64) LargeBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeLarge]byte)(ptr)
}

// GreatArrayFromSlice returns a GreatBuffer by copying from the slice at the given offset.
func GreatArrayFromSlice(s []byte, offset int64) GreatBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeGreat]byte)(ptr)
}

// HugeArrayFromSlice returns a HugeBuffer by copying from the slice at the given offset.
func HugeArrayFromSlice(s []byte, offset int64) HugeBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeHuge]byte)(ptr)
}

// VastArrayFromSlice returns a VastBuffer by copying from the slice at the given offset.
func VastArrayFromSlice(s []byte, offset int64) VastBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeVast]byte)(ptr)
}

// GiantArrayFromSlice returns a GiantBuffer by copying from the slice at the given offset.
func GiantArrayFromSlice(s []byte, offset int64) GiantBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeGiant]byte)(ptr)
}

// TitanArrayFromSlice returns a TitanBuffer by copying from the slice at the given offset.
func TitanArrayFromSlice(s []byte, offset int64) TitanBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeTitan]byte)(ptr)
}

// SliceOfPicoArray returns a slice of n PicoBuffers viewed from the underlying slice.
//
// The returned slice references the same memory as s[offset:]; modifications
// to either will be visible in both. The caller must ensure:
//   - offset + n*BufferSizePico <= len(s)
//   - n >= 1 (panics otherwise)
func SliceOfPicoArray(s []byte, offset int64, n int) []PicoBuffer {
	if n < 1 {
		panic("invalid array count")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*PicoBuffer)(unsafe.Add(base, offset)), n)
}

// SliceOfNanoArray returns a slice of NanoBuffer views of the underlying slice starting at offset.
func SliceOfNanoArray(s []byte, offset int64, n int) []NanoBuffer {
	if n < 1 {
		panic("invalid array count")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*NanoBuffer)(unsafe.Add(base, offset)), n)
}

// SliceOfMicroArray returns a slice of MicroBuffer views of the underlying slice starting at offset.
func SliceOfMicroArray(s []byte, offset int64, n int) []MicroBuffer {
	if n < 1 {
		panic("invalid array count")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*MicroBuffer)(unsafe.Add(base, offset)), n)
}

// SliceOfSmallArray returns a slice of SmallBuffer views of the underlying slice starting at offset.
func SliceOfSmallArray(s []byte, offset int64, n int) []SmallBuffer {
	if n < 1 {
		panic("invalid array count")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*SmallBuffer)(unsafe.Add(base, offset)), n)
}

// SliceOfMediumArray returns a slice of MediumBuffer views of the underlying slice starting at offset.
func SliceOfMediumArray(s []byte, offset int64, n int) []MediumBuffer {
	if n < 1 {
		panic("invalid array count")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*MediumBuffer)(unsafe.Add(base, offset)), n)
}

// SliceOfBigArray returns a slice of BigBuffer views of the underlying slice starting at offset.
func SliceOfBigArray(s []byte, offset int64, n int) []BigBuffer {
	if n < 1 {
		panic("invalid array count")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*BigBuffer)(unsafe.Add(base, offset)), n)
}

// SliceOfLargeArray returns a slice of LargeBuffer views of the underlying slice starting at offset.
func SliceOfLargeArray(s []byte, offset int64, n int) []LargeBuffer {
	if n < 1 {
		panic("invalid array count")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*LargeBuffer)(unsafe.Add(base, offset)), n)
}

// SliceOfGreatArray returns a slice of GreatBuffer views of the underlying slice starting at offset.
func SliceOfGreatArray(s []byte, offset int64, n int) []GreatBuffer {
	if n < 1 {
		panic("invalid array count")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*GreatBuffer)(unsafe.Add(base, offset)), n)
}

// SliceOfHugeArray returns a slice of HugeBuffer views of the underlying slice starting at offset.
func SliceOfHugeArray(s []byte, offset int64, n int) []HugeBuffer {
	if n < 1 {
		panic("invalid array count")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*HugeBuffer)(unsafe.Add(base, offset)), n)
}

// SliceOfVastArray returns a slice of VastBuffer views of the underlying slice starting at offset.
func SliceOfVastArray(s []byte, offset int64, n int) []VastBuffer {
	if n < 1 {
		panic("invalid array count")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*VastBuffer)(unsafe.Add(base, offset)), n)
}

// SliceOfGiantArray returns a slice of GiantBuffer views of the underlying slice starting at offset.
func SliceOfGiantArray(s []byte, offset int64, n int) []GiantBuffer {
	if n < 1 {
		panic("invalid array count")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*GiantBuffer)(unsafe.Add(base, offset)), n)
}

// SliceOfTitanArray returns a slice of TitanBuffer views of the underlying slice starting at offset.
func SliceOfTitanArray(s []byte, offset int64, n int) []TitanBuffer {
	if n < 1 {
		panic("invalid array count")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*TitanBuffer)(unsafe.Add(base, offset)), n)
}

// NewRegisterBufferPool creates a RegisterBufferPool for io_uring buffer registration.
//
// The actual capacity is rounded up to the next power of two.
// RegisterBuffer uses LargeBuffer size (128 KiB), suitable for io_uring provided buffers.
func NewRegisterBufferPool(capacity int) *RegisterBufferPool {
	return NewBoundedPool[RegisterBuffer](capacity)
}

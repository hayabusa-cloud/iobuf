// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package iobuf

import (
	"unsafe"
)

const (
	registerBufferSize = BufferSizeHuge
)

type (

	// RegisterBuffer represents a fixed-size buffer used for registering with the I/O ring.
	RegisterBuffer [registerBufferSize]byte

	// RegisterBufferPool represents a pool of fixed-size buffers used for registering with the I/O ring.
	RegisterBufferPool = BoundedPool[RegisterBuffer]
)

// AlignedMem returns a byte slice with the specified size
// and starting address aligned to the memory page size.
func AlignedMem(size int, pageSize uintptr) []byte {
	p := make([]byte, uintptr(size)+pageSize-1)
	ptr := uintptr(unsafe.Pointer(unsafe.SliceData(p)))
	ptr = ((ptr + pageSize - 1) / pageSize) * pageSize
	return unsafe.Slice((*byte)(unsafe.Pointer(ptr)), size)
}

// AlignedMemBlocks returns n bytes slices that
// have length with memory page size and address
// starts from multiple of memory page size
func AlignedMemBlocks(n int, pageSize uintptr) (blocks [][]byte) {
	if n < 1 {
		panic("bad block num")
	}
	blocks = make([][]byte, n)
	p := make([]byte, int(pageSize)*(n+1))
	ptr := uintptr(unsafe.Pointer(&p[0]))
	off := ptr - (ptr & ^(pageSize - 1))
	for i := range n {
		blocks[i] = unsafe.Slice(&p[uintptr(i)*pageSize-off], pageSize)
	}
	return
}

// AlignedMemBlock returns one aligned block with default page size.
func AlignedMemBlock() []byte {
	return AlignedMemBlocks(1, PageSize)[0]
}

// NewBuffers creates and initializes a new Buffers with a given n and size
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

// Buffer size tiers (bytes): Pico=16, Nano=64, Micro=256, Small=1K,
// Medium=4K, Large=16K, Huge=64K, Giant=256K.
const (
	_ = 1 << (iota * 2)
	_
	BufferSizePico   // 16 bytes
	BufferSizeNano   // 64 bytes
	BufferSizeMicro  // 256 bytes
	BufferSizeSmall  // 1 KiB
	BufferSizeMedium // 4 KiB
	BufferSizeLarge  // 16 KiB
	BufferSizeHuge   // 64 KiB
	BufferSizeGiant  // 256 KiB
)

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

// NewLargeBuffer returns a zero-initialized LargeBuffer.
func NewLargeBuffer() LargeBuffer { return LargeBuffer{} }

// NewHugeBuffer returns a zero-initialized HugeBuffer.
func NewHugeBuffer() HugeBuffer { return HugeBuffer{} }

// NewGiantBuffer returns a zero-initialized GiantBuffer.
func NewGiantBuffer() GiantBuffer { return GiantBuffer{} }

// BufferType is a type constraint for tiered buffer types.
type BufferType interface {
	PicoBuffer | NanoBuffer | MicroBuffer | SmallBuffer | MediumBuffer | LargeBuffer | HugeBuffer | GiantBuffer
}

type (
	// PicoBuffer represents a fixed-size byte array of 16 bytes.
	PicoBuffer [BufferSizePico]byte

	// NanoBuffer represents a fixed-size byte array of 64 bytes.
	NanoBuffer [BufferSizeNano]byte

	// MicroBuffer represents a fixed-size byte array of 256 bytes.
	MicroBuffer [BufferSizeMicro]byte

	// SmallBuffer represents a fixed-size byte array of 1,024 bytes (1 KiB).
	SmallBuffer [BufferSizeSmall]byte

	// MediumBuffer represents a fixed-size byte array of 4,096 bytes (4 KiB).
	MediumBuffer [BufferSizeMedium]byte

	// LargeBuffer represents a fixed-size byte array of 16,384 bytes (16 KiB).
	LargeBuffer [BufferSizeLarge]byte

	// HugeBuffer represents a fixed-size byte array of 65,536 bytes (64 KiB).
	HugeBuffer [BufferSizeHuge]byte

	// GiantBuffer represents a fixed-size byte array of 262,144 bytes (256 KiB).
	GiantBuffer [BufferSizeGiant]byte
)

// Reset is a no-op implementation satisfying the Pool item contract.
func (b PicoBuffer) Reset()   {}
func (b NanoBuffer) Reset()   {}
func (b MicroBuffer) Reset()  {}
func (b SmallBuffer) Reset()  {}
func (b MediumBuffer) Reset() {}
func (b LargeBuffer) Reset()  {}
func (b HugeBuffer) Reset()   {}
func (b GiantBuffer) Reset()  {}

// PicoArrayFromSlice returns a PicoBuffer view of the underlying slice at the given offset.
func PicoArrayFromSlice(s []byte, offset int64) PicoBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizePico]byte)(ptr)
}

// NanoArrayFromSlice returns a NanoBuffer view of the underlying slice at the given offset.
func NanoArrayFromSlice(s []byte, offset int64) NanoBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeNano]byte)(ptr)
}

// MicroArrayFromSlice returns a MicroBuffer view of the underlying slice at the given offset.
func MicroArrayFromSlice(s []byte, offset int64) MicroBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeMicro]byte)(ptr)
}

// SmallArrayFromSlice returns a SmallBuffer view of the underlying slice at the given offset.
func SmallArrayFromSlice(s []byte, offset int64) SmallBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeSmall]byte)(ptr)
}

// MediumArrayFromSlice returns a MediumBuffer view of the underlying slice at the given offset.
func MediumArrayFromSlice(s []byte, offset int64) MediumBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeMedium]byte)(ptr)
}

// LargeArrayFromSlice returns a LargeBuffer view of the underlying slice at the given offset.
func LargeArrayFromSlice(s []byte, offset int64) LargeBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeLarge]byte)(ptr)
}

// HugeArrayFromSlice returns a HugeBuffer view of the underlying slice at the given offset.
func HugeArrayFromSlice(s []byte, offset int64) HugeBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeHuge]byte)(ptr)
}

// GiantArrayFromSlice returns a GiantBuffer view of the underlying slice at the given offset.
func GiantArrayFromSlice(s []byte, offset int64) GiantBuffer {
	ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(s)), offset)
	return *(*[BufferSizeGiant]byte)(ptr)
}

// SliceOfPicoArray returns a slice of PicoBuffer views of the underlying slice starting at offset.
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

// SliceOfLargeArray returns a slice of LargeBuffer views of the underlying slice starting at offset.
func SliceOfLargeArray(s []byte, offset int64, n int) []LargeBuffer {
	if n < 1 {
		panic("invalid array count")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*LargeBuffer)(unsafe.Add(base, offset)), n)
}

// SliceOfHugeArray returns a slice of HugeBuffer views of the underlying slice starting at offset.
func SliceOfHugeArray(s []byte, offset int64, n int) []HugeBuffer {
	if n < 1 {
		panic("invalid array count")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*HugeBuffer)(unsafe.Add(base, offset)), n)
}

// SliceOfGiantArray returns a slice of GiantBuffer views of the underlying slice starting at offset.
func SliceOfGiantArray(s []byte, offset int64, n int) []GiantBuffer {
	if n < 1 {
		panic("invalid array count")
	}
	base := unsafe.Pointer(unsafe.SliceData(s))
	return unsafe.Slice((*GiantBuffer)(unsafe.Add(base, offset)), n)
}

// NewRegisterBufferPool creates a new instance of RegisterBufferPool with the specified capacity.
func NewRegisterBufferPool(capacity int) *RegisterBufferPool {
	return NewBoundedPool[RegisterBuffer](capacity)
}

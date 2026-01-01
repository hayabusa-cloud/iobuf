// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package iobuf

import (
	"unsafe"
)

// IoVec represents a scatter/gather I/O descriptor compatible with the
// standard Linux struct iovec. It is used to pass multiple non-contiguous
// user-space buffers to the kernel in a single vectored I/O system call
// (readv, writev, preadv, pwritev, io_uring operations).
//
// Memory layout matches the C struct iovec exactly:
//
//	struct iovec {
//	    void  *iov_base;  // Starting address
//	    size_t iov_len;   // Number of bytes
//	};
//
// The caller must ensure Base points to valid memory for the lifetime of
// any I/O operation using this IoVec.
type IoVec struct {
	Base *byte  // Starting address of the memory block
	Len  uint64 // Number of bytes to transfer
}

// IoVecFromBytesSlice converts a slice of byte slices to a pointer and count
// suitable for io_uring buffer registration (IORING_REGISTER_BUFFERS2).
// Returns the address of the first IoVec element and the number of elements.
//
// Note: The returned address points to a newly allocated []IoVec slice.
// The caller must ensure the input slices remain valid for the lifetime
// of the registration.
func IoVecFromBytesSlice(iov [][]byte) (addr uintptr, n int) {
	if len(iov) == 0 {
		return 0, 0
	}
	vec := make([]IoVec, len(iov))
	for i := range len(iov) {
		vec[i] = IoVec{Base: unsafe.SliceData(iov[i]), Len: uint64(len(iov[i]))}
	}
	addr, n = uintptr(unsafe.Pointer(unsafe.SliceData(vec))), len(vec)
	return
}

// IoVecAddrLen extracts the raw pointer and length from an IoVec slice
// for direct syscall consumption (readv, writev, io_uring submission).
//
// Returns (0, 0) for empty or nil slices.
func IoVecAddrLen(vec []IoVec) (addr uintptr, n int) {
	if len(vec) == 0 {
		return 0, 0
	}
	addr, n = uintptr(unsafe.Pointer(unsafe.SliceData(vec))), len(vec)
	return
}

// IoVecFromPicoBuffers converts a slice of PicoBuffer to an IoVec slice.
// The returned IoVec elements point directly to the buffer memory without copying.
func IoVecFromPicoBuffers(buffers []PicoBuffer) []IoVec {
	if len(buffers) == 0 {
		return nil
	}
	vec := make([]IoVec, len(buffers))
	for i := range len(buffers) {
		vec[i] = IoVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizePico}
	}
	return vec
}

// IoVecFromNanoBuffers converts a slice of NanoBuffer to an IoVec slice.
// The returned IoVec elements point directly to the buffer memory without copying.
func IoVecFromNanoBuffers(buffers []NanoBuffer) []IoVec {
	if len(buffers) == 0 {
		return nil
	}
	vec := make([]IoVec, len(buffers))
	for i := range len(buffers) {
		vec[i] = IoVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeNano}
	}
	return vec
}

// IoVecFromMicroBuffers converts a slice of MicroBuffer to an IoVec slice.
// The returned IoVec elements point directly to the buffer memory without copying.
func IoVecFromMicroBuffers(buffers []MicroBuffer) []IoVec {
	if len(buffers) == 0 {
		return nil
	}
	vec := make([]IoVec, len(buffers))
	for i := range len(buffers) {
		vec[i] = IoVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeMicro}
	}
	return vec
}

// IoVecFromSmallBuffers converts a slice of SmallBuffer to an IoVec slice.
// The returned IoVec elements point directly to the buffer memory without copying.
func IoVecFromSmallBuffers(buffers []SmallBuffer) []IoVec {
	if len(buffers) == 0 {
		return nil
	}
	vec := make([]IoVec, len(buffers))
	for i := range len(buffers) {
		vec[i] = IoVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeSmall}
	}
	return vec
}

// IoVecFromMediumBuffers converts a slice of MediumBuffer to an IoVec slice.
// The returned IoVec elements point directly to the buffer memory without copying.
func IoVecFromMediumBuffers(buffers []MediumBuffer) []IoVec {
	if len(buffers) == 0 {
		return nil
	}
	vec := make([]IoVec, len(buffers))
	for i := range len(buffers) {
		vec[i] = IoVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeMedium}
	}
	return vec
}

// IoVecFromBigBuffers converts a slice of BigBuffer to an IoVec slice.
// The returned IoVec elements point directly to the buffer memory without copying.
func IoVecFromBigBuffers(buffers []BigBuffer) []IoVec {
	if len(buffers) == 0 {
		return nil
	}
	vec := make([]IoVec, len(buffers))
	for i := range len(buffers) {
		vec[i] = IoVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeBig}
	}
	return vec
}

// IoVecFromLargeBuffers converts a slice of LargeBuffer to an IoVec slice.
// The returned IoVec elements point directly to the buffer memory without copying.
func IoVecFromLargeBuffers(buffers []LargeBuffer) []IoVec {
	if len(buffers) == 0 {
		return nil
	}
	vec := make([]IoVec, len(buffers))
	for i := range len(buffers) {
		vec[i] = IoVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeLarge}
	}
	return vec
}

// IoVecFromGreatBuffers converts a slice of GreatBuffer to an IoVec slice.
// The returned IoVec elements point directly to the buffer memory without copying.
func IoVecFromGreatBuffers(buffers []GreatBuffer) []IoVec {
	if len(buffers) == 0 {
		return nil
	}
	vec := make([]IoVec, len(buffers))
	for i := range len(buffers) {
		vec[i] = IoVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeGreat}
	}
	return vec
}

// IoVecFromHugeBuffers converts a slice of HugeBuffer to an IoVec slice.
// The returned IoVec elements point directly to the buffer memory without copying.
func IoVecFromHugeBuffers(buffers []HugeBuffer) []IoVec {
	if len(buffers) == 0 {
		return nil
	}
	vec := make([]IoVec, len(buffers))
	for i := range len(buffers) {
		vec[i] = IoVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeHuge}
	}
	return vec
}

// IoVecFromVastBuffers converts a slice of VastBuffer to an IoVec slice.
// The returned IoVec elements point directly to the buffer memory without copying.
func IoVecFromVastBuffers(buffers []VastBuffer) []IoVec {
	if len(buffers) == 0 {
		return nil
	}
	vec := make([]IoVec, len(buffers))
	for i := range len(buffers) {
		vec[i] = IoVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeVast}
	}
	return vec
}

// IoVecFromGiantBuffers converts a slice of GiantBuffer to an IoVec slice.
// The returned IoVec elements point directly to the buffer memory without copying.
func IoVecFromGiantBuffers(buffers []GiantBuffer) []IoVec {
	if len(buffers) == 0 {
		return nil
	}
	vec := make([]IoVec, len(buffers))
	for i := range len(buffers) {
		vec[i] = IoVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeGiant}
	}
	return vec
}

// IoVecFromTitanBuffers converts a slice of TitanBuffer to an IoVec slice.
// The returned IoVec elements point directly to the buffer memory without copying.
func IoVecFromTitanBuffers(buffers []TitanBuffer) []IoVec {
	if len(buffers) == 0 {
		return nil
	}
	vec := make([]IoVec, len(buffers))
	for i := range len(buffers) {
		vec[i] = IoVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: BufferSizeTitan}
	}
	return vec
}

// IoVecFromRegisteredBuffers converts a slice of RegisterBuffer to an IoVec slice.
// The returned IoVec elements point directly to the buffer memory without copying.
func IoVecFromRegisteredBuffers(buffers []RegisterBuffer) []IoVec {
	if len(buffers) == 0 {
		return nil
	}
	vec := make([]IoVec, len(buffers))
	for i := range len(buffers) {
		vec[i] = IoVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: registerBufferSize}
	}
	return vec
}

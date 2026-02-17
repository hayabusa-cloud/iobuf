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
	for i := range iov {
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

// IoVecFrom converts a slice of typed buffers to an IoVec slice.
// The returned IoVec elements point directly to the buffer memory without copying.
func IoVecFrom[T BufferType](buffers []T) []IoVec {
	if len(buffers) == 0 {
		return nil
	}
	vec := make([]IoVec, len(buffers))
	for i := range buffers {
		vec[i] = IoVec{
			Base: (*byte)(unsafe.Pointer(&buffers[i])),
			Len:  uint64(unsafe.Sizeof(buffers[i])),
		}
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
	for i := range buffers {
		vec[i] = IoVec{Base: (*byte)(unsafe.Pointer(&buffers[i])), Len: registerBufferSize}
	}
	return vec
}

// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package iobuf_test

import (
	"testing"
	"unsafe"

	"code.hybscloud.com/iobuf"
)

func TestAlignedMem_PageAlignment(t *testing.T) {
	const size = 8192
	mem := iobuf.AlignedMem(size, iobuf.PageSize)

	if len(mem) != size {
		t.Errorf("AlignedMem length = %d, want %d", len(mem), size)
	}

	ptr := uintptr(unsafe.Pointer(unsafe.SliceData(mem)))
	if ptr%iobuf.PageSize != 0 {
		t.Errorf("AlignedMem not page-aligned: address %#x %% %d = %d", ptr, iobuf.PageSize, ptr%iobuf.PageSize)
	}
}

func TestAlignedMem_SmallAllocation(t *testing.T) {
	const size = 64
	mem := iobuf.AlignedMem(size, iobuf.PageSize)

	if len(mem) != size {
		t.Errorf("AlignedMem length = %d, want %d", len(mem), size)
	}

	ptr := uintptr(unsafe.Pointer(unsafe.SliceData(mem)))
	if ptr%iobuf.PageSize != 0 {
		t.Errorf("AlignedMem not page-aligned: address %#x %% %d = %d", ptr, iobuf.PageSize, ptr%iobuf.PageSize)
	}
}

func TestAlignedMemBlocks(t *testing.T) {
	const n = 4
	blocks := iobuf.AlignedMemBlocks(n, iobuf.PageSize)

	if len(blocks) != n {
		t.Errorf("AlignedMemBlocks returned %d blocks, want %d", len(blocks), n)
	}

	for i, block := range blocks {
		if uintptr(len(block)) != iobuf.PageSize {
			t.Errorf("block[%d] length = %d, want %d", i, len(block), iobuf.PageSize)
		}
		ptr := uintptr(unsafe.Pointer(unsafe.SliceData(block)))
		if ptr%iobuf.PageSize != 0 {
			t.Errorf("block[%d] not page-aligned: address %#x %% %d = %d", i, ptr, iobuf.PageSize, ptr%iobuf.PageSize)
		}
	}
}

func TestAlignedMemBlock(t *testing.T) {
	block := iobuf.AlignedMemBlock()

	if uintptr(len(block)) != iobuf.PageSize {
		t.Errorf("AlignedMemBlock length = %d, want %d", len(block), iobuf.PageSize)
	}

	ptr := uintptr(unsafe.Pointer(unsafe.SliceData(block)))
	if ptr%iobuf.PageSize != 0 {
		t.Errorf("AlignedMemBlock not page-aligned: address %#x %% %d = %d", ptr, iobuf.PageSize, ptr%iobuf.PageSize)
	}
}

func TestBufferSizes(t *testing.T) {
	// Verify buffer sizes follow the expected pattern (powers of 4, starting at 16)
	expectedSizes := []int{
		16,     // Pico: 4^2
		64,     // Nano: 4^3
		256,    // Micro: 4^4
		1024,   // Small: 4^5
		4096,   // Medium: 4^6
		16384,  // Large: 4^7
		65536,  // Huge: 4^8
		262144, // Giant: 4^9
	}

	actualSizes := []int{
		iobuf.BufferSizePico,
		iobuf.BufferSizeNano,
		iobuf.BufferSizeMicro,
		iobuf.BufferSizeSmall,
		iobuf.BufferSizeMedium,
		iobuf.BufferSizeLarge,
		iobuf.BufferSizeHuge,
		iobuf.BufferSizeGiant,
	}

	for i, expected := range expectedSizes {
		if actualSizes[i] != expected {
			t.Errorf("buffer size[%d] = %d, want %d", i, actualSizes[i], expected)
		}
	}
}

func TestNewBuffers(t *testing.T) {
	const n, size = 8, 256
	bufs := iobuf.NewBuffers(n, size)

	if len(bufs) != n {
		t.Errorf("NewBuffers returned %d buffers, want %d", len(bufs), n)
	}

	for i, buf := range bufs {
		if len(buf) != size {
			t.Errorf("buffer[%d] length = %d, want %d", i, len(buf), size)
		}
	}
}

func TestNewBuffers_ZeroSize(t *testing.T) {
	const n = 4
	bufs := iobuf.NewBuffers(n, 0)

	if len(bufs) != n {
		t.Errorf("NewBuffers returned %d buffers, want %d", len(bufs), n)
	}

	for i, buf := range bufs {
		if len(buf) != 0 {
			t.Errorf("buffer[%d] length = %d, want 0", i, len(buf))
		}
	}
}

func TestRegisterBufferPool(t *testing.T) {
	const capacity = 16
	pool := iobuf.NewRegisterBufferPool(capacity)

	if pool.Cap() != capacity {
		t.Errorf("RegisterBufferPool capacity = %d, want %d", pool.Cap(), capacity)
	}
}

func TestNewBuffers_InvalidN(t *testing.T) {
	bufs := iobuf.NewBuffers(0, 64)
	if len(bufs) != 0 {
		t.Errorf("NewBuffers(0, 64) returned %d buffers, want 0", len(bufs))
	}

	bufs = iobuf.NewBuffers(-1, 64)
	if len(bufs) != 0 {
		t.Errorf("NewBuffers(-1, 64) returned %d buffers, want 0", len(bufs))
	}
}

func TestAlignedMemBlocks_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("AlignedMemBlocks(0, PageSize) did not panic")
		}
	}()
	_ = iobuf.AlignedMemBlocks(0, iobuf.PageSize)
}

func TestAlignedMem_NonStandardPageSize(t *testing.T) {
	const customPageSize = 8192
	const size = 16384
	mem := iobuf.AlignedMem(size, customPageSize)

	if len(mem) != size {
		t.Errorf("AlignedMem length = %d, want %d", len(mem), size)
	}

	ptr := uintptr(unsafe.Pointer(unsafe.SliceData(mem)))
	if ptr%customPageSize != 0 {
		t.Errorf("AlignedMem not aligned to %d: address %#x %% %d = %d",
			customPageSize, ptr, customPageSize, ptr%customPageSize)
	}
}

func TestSetPageSize(t *testing.T) {
	original := iobuf.PageSize
	defer iobuf.SetPageSize(int(original))

	iobuf.SetPageSize(8192)
	if iobuf.PageSize != 8192 {
		t.Errorf("SetPageSize(8192) resulted in PageSize = %d, want 8192", iobuf.PageSize)
	}
}

func TestNewTierBuffers(t *testing.T) {
	t.Run("NewPicoBuffer", func(t *testing.T) {
		buf := iobuf.NewPicoBuffer()
		if len(buf) != iobuf.BufferSizePico {
			t.Errorf("NewPicoBuffer size = %d, want %d", len(buf), iobuf.BufferSizePico)
		}
	})

	t.Run("NewNanoBuffer", func(t *testing.T) {
		buf := iobuf.NewNanoBuffer()
		if len(buf) != iobuf.BufferSizeNano {
			t.Errorf("NewNanoBuffer size = %d, want %d", len(buf), iobuf.BufferSizeNano)
		}
	})

	t.Run("NewMicroBuffer", func(t *testing.T) {
		buf := iobuf.NewMicroBuffer()
		if len(buf) != iobuf.BufferSizeMicro {
			t.Errorf("NewMicroBuffer size = %d, want %d", len(buf), iobuf.BufferSizeMicro)
		}
	})

	t.Run("NewSmallBuffer", func(t *testing.T) {
		buf := iobuf.NewSmallBuffer()
		if len(buf) != iobuf.BufferSizeSmall {
			t.Errorf("NewSmallBuffer size = %d, want %d", len(buf), iobuf.BufferSizeSmall)
		}
	})

	t.Run("NewMediumBuffer", func(t *testing.T) {
		buf := iobuf.NewMediumBuffer()
		if len(buf) != iobuf.BufferSizeMedium {
			t.Errorf("NewMediumBuffer size = %d, want %d", len(buf), iobuf.BufferSizeMedium)
		}
	})

	t.Run("NewLargeBuffer", func(t *testing.T) {
		buf := iobuf.NewLargeBuffer()
		if len(buf) != iobuf.BufferSizeLarge {
			t.Errorf("NewLargeBuffer size = %d, want %d", len(buf), iobuf.BufferSizeLarge)
		}
	})

	t.Run("NewHugeBuffer", func(t *testing.T) {
		buf := iobuf.NewHugeBuffer()
		if len(buf) != iobuf.BufferSizeHuge {
			t.Errorf("NewHugeBuffer size = %d, want %d", len(buf), iobuf.BufferSizeHuge)
		}
	})

	t.Run("NewGiantBuffer", func(t *testing.T) {
		buf := iobuf.NewGiantBuffer()
		if len(buf) != iobuf.BufferSizeGiant {
			t.Errorf("NewGiantBuffer size = %d, want %d", len(buf), iobuf.BufferSizeGiant)
		}
	})
}

func TestBufferReset(t *testing.T) {
	t.Run("PicoBuffer", func(t *testing.T) {
		buf := iobuf.PicoBuffer{}
		buf[0] = 0xFF
		buf.Reset()
		if buf[0] != 0xFF {
			t.Error("Reset() should be a no-op, but modified buffer")
		}
	})

	t.Run("NanoBuffer", func(t *testing.T) {
		buf := iobuf.NanoBuffer{}
		buf[0] = 0xFF
		buf.Reset()
		if buf[0] != 0xFF {
			t.Error("Reset() should be a no-op, but modified buffer")
		}
	})

	t.Run("MicroBuffer", func(t *testing.T) {
		buf := iobuf.MicroBuffer{}
		buf[0] = 0xFF
		buf.Reset()
		if buf[0] != 0xFF {
			t.Error("Reset() should be a no-op, but modified buffer")
		}
	})

	t.Run("SmallBuffer", func(t *testing.T) {
		buf := iobuf.SmallBuffer{}
		buf[0] = 0xFF
		buf.Reset()
		if buf[0] != 0xFF {
			t.Error("Reset() should be a no-op, but modified buffer")
		}
	})

	t.Run("MediumBuffer", func(t *testing.T) {
		buf := iobuf.MediumBuffer{}
		buf[0] = 0xFF
		buf.Reset()
		if buf[0] != 0xFF {
			t.Error("Reset() should be a no-op, but modified buffer")
		}
	})

	t.Run("LargeBuffer", func(t *testing.T) {
		buf := iobuf.LargeBuffer{}
		buf[0] = 0xFF
		buf.Reset()
		if buf[0] != 0xFF {
			t.Error("Reset() should be a no-op, but modified buffer")
		}
	})

	t.Run("HugeBuffer", func(t *testing.T) {
		buf := iobuf.HugeBuffer{}
		buf[0] = 0xFF
		buf.Reset()
		if buf[0] != 0xFF {
			t.Error("Reset() should be a no-op, but modified buffer")
		}
	})

	t.Run("GiantBuffer", func(t *testing.T) {
		buf := iobuf.GiantBuffer{}
		buf[0] = 0xFF
		buf.Reset()
		if buf[0] != 0xFF {
			t.Error("Reset() should be a no-op, but modified buffer")
		}
	})
}

func TestArrayFromSlice(t *testing.T) {
	data := make([]byte, iobuf.BufferSizeGiant*2)
	for i := range data {
		data[i] = byte(i % 256)
	}

	t.Run("PicoArrayFromSlice", func(t *testing.T) {
		arr := iobuf.PicoArrayFromSlice(data, 0)
		if arr[0] != data[0] {
			t.Errorf("PicoArrayFromSlice[0] = %d, want %d", arr[0], data[0])
		}
		arr2 := iobuf.PicoArrayFromSlice(data, 16)
		if arr2[0] != data[16] {
			t.Errorf("PicoArrayFromSlice offset 16 [0] = %d, want %d", arr2[0], data[16])
		}
	})

	t.Run("NanoArrayFromSlice", func(t *testing.T) {
		arr := iobuf.NanoArrayFromSlice(data, 0)
		if arr[0] != data[0] {
			t.Errorf("NanoArrayFromSlice[0] = %d, want %d", arr[0], data[0])
		}
	})

	t.Run("MicroArrayFromSlice", func(t *testing.T) {
		arr := iobuf.MicroArrayFromSlice(data, 0)
		if arr[0] != data[0] {
			t.Errorf("MicroArrayFromSlice[0] = %d, want %d", arr[0], data[0])
		}
	})

	t.Run("SmallArrayFromSlice", func(t *testing.T) {
		arr := iobuf.SmallArrayFromSlice(data, 0)
		if arr[0] != data[0] {
			t.Errorf("SmallArrayFromSlice[0] = %d, want %d", arr[0], data[0])
		}
	})

	t.Run("MediumArrayFromSlice", func(t *testing.T) {
		arr := iobuf.MediumArrayFromSlice(data, 0)
		if arr[0] != data[0] {
			t.Errorf("MediumArrayFromSlice[0] = %d, want %d", arr[0], data[0])
		}
	})

	t.Run("LargeArrayFromSlice", func(t *testing.T) {
		arr := iobuf.LargeArrayFromSlice(data, 0)
		if arr[0] != data[0] {
			t.Errorf("LargeArrayFromSlice[0] = %d, want %d", arr[0], data[0])
		}
	})

	t.Run("HugeArrayFromSlice", func(t *testing.T) {
		arr := iobuf.HugeArrayFromSlice(data, 0)
		if arr[0] != data[0] {
			t.Errorf("HugeArrayFromSlice[0] = %d, want %d", arr[0], data[0])
		}
	})

	t.Run("GiantArrayFromSlice", func(t *testing.T) {
		arr := iobuf.GiantArrayFromSlice(data, 0)
		if arr[0] != data[0] {
			t.Errorf("GiantArrayFromSlice[0] = %d, want %d", arr[0], data[0])
		}
	})
}

func TestSliceOfArray(t *testing.T) {
	data := make([]byte, iobuf.BufferSizeGiant*4)
	for i := range data {
		data[i] = byte(i % 256)
	}

	t.Run("SliceOfPicoArray", func(t *testing.T) {
		arr := iobuf.SliceOfPicoArray(data, 0, 4)
		if len(arr) != 4 {
			t.Errorf("SliceOfPicoArray len = %d, want 4", len(arr))
		}
		if arr[0][0] != data[0] {
			t.Errorf("SliceOfPicoArray[0][0] = %d, want %d", arr[0][0], data[0])
		}
	})

	t.Run("SliceOfNanoArray", func(t *testing.T) {
		arr := iobuf.SliceOfNanoArray(data, 0, 4)
		if len(arr) != 4 {
			t.Errorf("SliceOfNanoArray len = %d, want 4", len(arr))
		}
	})

	t.Run("SliceOfMicroArray", func(t *testing.T) {
		arr := iobuf.SliceOfMicroArray(data, 0, 4)
		if len(arr) != 4 {
			t.Errorf("SliceOfMicroArray len = %d, want 4", len(arr))
		}
	})

	t.Run("SliceOfSmallArray", func(t *testing.T) {
		arr := iobuf.SliceOfSmallArray(data, 0, 4)
		if len(arr) != 4 {
			t.Errorf("SliceOfSmallArray len = %d, want 4", len(arr))
		}
	})

	t.Run("SliceOfMediumArray", func(t *testing.T) {
		arr := iobuf.SliceOfMediumArray(data, 0, 4)
		if len(arr) != 4 {
			t.Errorf("SliceOfMediumArray len = %d, want 4", len(arr))
		}
	})

	t.Run("SliceOfLargeArray", func(t *testing.T) {
		arr := iobuf.SliceOfLargeArray(data, 0, 4)
		if len(arr) != 4 {
			t.Errorf("SliceOfLargeArray len = %d, want 4", len(arr))
		}
	})

	t.Run("SliceOfHugeArray", func(t *testing.T) {
		arr := iobuf.SliceOfHugeArray(data, 0, 4)
		if len(arr) != 4 {
			t.Errorf("SliceOfHugeArray len = %d, want 4", len(arr))
		}
	})

	t.Run("SliceOfGiantArray", func(t *testing.T) {
		arr := iobuf.SliceOfGiantArray(data, 0, 4)
		if len(arr) != 4 {
			t.Errorf("SliceOfGiantArray len = %d, want 4", len(arr))
		}
	})
}

func TestSliceOfArray_Panic(t *testing.T) {
	data := make([]byte, 1024)

	testCases := []struct {
		name string
		fn   func()
	}{
		{"SliceOfPicoArray_n0", func() { iobuf.SliceOfPicoArray(data, 0, 0) }},
		{"SliceOfPicoArray_nNeg", func() { iobuf.SliceOfPicoArray(data, 0, -1) }},
		{"SliceOfNanoArray_n0", func() { iobuf.SliceOfNanoArray(data, 0, 0) }},
		{"SliceOfMicroArray_n0", func() { iobuf.SliceOfMicroArray(data, 0, 0) }},
		{"SliceOfSmallArray_n0", func() { iobuf.SliceOfSmallArray(data, 0, 0) }},
		{"SliceOfMediumArray_n0", func() { iobuf.SliceOfMediumArray(data, 0, 0) }},
		{"SliceOfLargeArray_n0", func() { iobuf.SliceOfLargeArray(data, 0, 0) }},
		{"SliceOfHugeArray_n0", func() { iobuf.SliceOfHugeArray(data, 0, 0) }},
		{"SliceOfGiantArray_n0", func() { iobuf.SliceOfGiantArray(data, 0, 0) }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("%s did not panic", tc.name)
				}
			}()
			tc.fn()
		})
	}
}

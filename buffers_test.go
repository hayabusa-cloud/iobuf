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
	// Verify buffer sizes follow the expected pattern (powers of 4, starting at 32)
	expectedSizes := []int{
		32,        // Pico: 2^5
		128,       // Nano: 2^7
		512,       // Micro: 2^9
		2048,      // Small: 2^11
		8192,      // Medium: 2^13
		32768,     // Big: 2^15
		131072,    // Large: 2^17
		524288,    // Great: 2^19
		2097152,   // Huge: 2^21
		8388608,   // Vast: 2^23
		33554432,  // Giant: 2^25
		134217728, // Titan: 2^27
	}

	actualSizes := []int{
		iobuf.BufferSizePico,
		iobuf.BufferSizeNano,
		iobuf.BufferSizeMicro,
		iobuf.BufferSizeSmall,
		iobuf.BufferSizeMedium,
		iobuf.BufferSizeBig,
		iobuf.BufferSizeLarge,
		iobuf.BufferSizeGreat,
		iobuf.BufferSizeHuge,
		iobuf.BufferSizeVast,
		iobuf.BufferSizeGiant,
		iobuf.BufferSizeTitan,
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

	t.Run("NewBigBuffer", func(t *testing.T) {
		buf := iobuf.NewBigBuffer()
		if len(buf) != iobuf.BufferSizeBig {
			t.Errorf("NewBigBuffer size = %d, want %d", len(buf), iobuf.BufferSizeBig)
		}
	})

	t.Run("NewLargeBuffer", func(t *testing.T) {
		buf := iobuf.NewLargeBuffer()
		if len(buf) != iobuf.BufferSizeLarge {
			t.Errorf("NewLargeBuffer size = %d, want %d", len(buf), iobuf.BufferSizeLarge)
		}
	})

	t.Run("NewGreatBuffer", func(t *testing.T) {
		buf := iobuf.NewGreatBuffer()
		if len(buf) != iobuf.BufferSizeGreat {
			t.Errorf("NewGreatBuffer size = %d, want %d", len(buf), iobuf.BufferSizeGreat)
		}
	})

	t.Run("NewHugeBuffer", func(t *testing.T) {
		buf := iobuf.NewHugeBuffer()
		if len(buf) != iobuf.BufferSizeHuge {
			t.Errorf("NewHugeBuffer size = %d, want %d", len(buf), iobuf.BufferSizeHuge)
		}
	})

	t.Run("NewVastBuffer", func(t *testing.T) {
		buf := iobuf.NewVastBuffer()
		if len(buf) != iobuf.BufferSizeVast {
			t.Errorf("NewVastBuffer size = %d, want %d", len(buf), iobuf.BufferSizeVast)
		}
	})

	t.Run("NewGiantBuffer", func(t *testing.T) {
		buf := iobuf.NewGiantBuffer()
		if len(buf) != iobuf.BufferSizeGiant {
			t.Errorf("NewGiantBuffer size = %d, want %d", len(buf), iobuf.BufferSizeGiant)
		}
	})

	t.Run("NewTitanBuffer", func(t *testing.T) {
		buf := iobuf.NewTitanBuffer()
		if len(buf) != iobuf.BufferSizeTitan {
			t.Errorf("NewTitanBuffer size = %d, want %d", len(buf), iobuf.BufferSizeTitan)
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

	t.Run("BigBuffer", func(t *testing.T) {
		buf := iobuf.BigBuffer{}
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

	t.Run("GreatBuffer", func(t *testing.T) {
		buf := iobuf.GreatBuffer{}
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

	t.Run("VastBuffer", func(t *testing.T) {
		buf := iobuf.VastBuffer{}
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

	t.Run("TitanBuffer", func(t *testing.T) {
		buf := iobuf.TitanBuffer{}
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

	t.Run("BigArrayFromSlice", func(t *testing.T) {
		arr := iobuf.BigArrayFromSlice(data, 0)
		if arr[0] != data[0] {
			t.Errorf("BigArrayFromSlice[0] = %d, want %d", arr[0], data[0])
		}
	})

	t.Run("GreatArrayFromSlice", func(t *testing.T) {
		arr := iobuf.GreatArrayFromSlice(data, 0)
		if arr[0] != data[0] {
			t.Errorf("GreatArrayFromSlice[0] = %d, want %d", arr[0], data[0])
		}
	})

	t.Run("VastArrayFromSlice", func(t *testing.T) {
		arr := iobuf.VastArrayFromSlice(data, 0)
		if arr[0] != data[0] {
			t.Errorf("VastArrayFromSlice[0] = %d, want %d", arr[0], data[0])
		}
	})

	t.Run("TitanArrayFromSlice", func(t *testing.T) {
		data := make([]byte, iobuf.BufferSizeTitan*2)
		data[0] = 0xAB
		data[iobuf.BufferSizeTitan] = 0xCD
		arr := iobuf.TitanArrayFromSlice(data, 0)
		if arr[0] != 0xAB {
			t.Errorf("TitanArrayFromSlice[0] = %d, want 0xAB", arr[0])
		}
		arr2 := iobuf.TitanArrayFromSlice(data, iobuf.BufferSizeTitan)
		if arr2[0] != 0xCD {
			t.Errorf("TitanArrayFromSlice offset [0] = %d, want 0xCD", arr2[0])
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

	t.Run("SliceOfBigArray", func(t *testing.T) {
		arr := iobuf.SliceOfBigArray(data, 0, 4)
		if len(arr) != 4 {
			t.Errorf("SliceOfBigArray len = %d, want 4", len(arr))
		}
	})

	t.Run("SliceOfGreatArray", func(t *testing.T) {
		arr := iobuf.SliceOfGreatArray(data, 0, 4)
		if len(arr) != 4 {
			t.Errorf("SliceOfGreatArray len = %d, want 4", len(arr))
		}
	})

	t.Run("SliceOfVastArray", func(t *testing.T) {
		arr := iobuf.SliceOfVastArray(data, 0, 4)
		if len(arr) != 4 {
			t.Errorf("SliceOfVastArray len = %d, want 4", len(arr))
		}
	})

	t.Run("SliceOfTitanArray", func(t *testing.T) {
		data := make([]byte, iobuf.BufferSizeTitan*2)
		data[0] = 0xAB
		data[iobuf.BufferSizeTitan] = 0xCD
		arr := iobuf.SliceOfTitanArray(data, 0, 2)
		if len(arr) != 2 {
			t.Errorf("SliceOfTitanArray len = %d, want 2", len(arr))
		}
		if arr[0][0] != 0xAB {
			t.Errorf("SliceOfTitanArray[0][0] = %d, want 0xAB", arr[0][0])
		}
		if arr[1][0] != 0xCD {
			t.Errorf("SliceOfTitanArray[1][0] = %d, want 0xCD", arr[1][0])
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

func TestTierBySize(t *testing.T) {
	tests := []struct {
		size     int
		expected iobuf.BufferTier
	}{
		// Exact boundaries (new sizes: power-of-4 from 32)
		{0, iobuf.TierPico},
		{1, iobuf.TierPico},
		{32, iobuf.TierPico},
		{33, iobuf.TierNano},
		{128, iobuf.TierNano},
		{129, iobuf.TierMicro},
		{512, iobuf.TierMicro},
		{513, iobuf.TierSmall},
		{2048, iobuf.TierSmall},
		{2049, iobuf.TierMedium},
		{8192, iobuf.TierMedium},
		{8193, iobuf.TierBig},
		{32768, iobuf.TierBig},
		{32769, iobuf.TierLarge},
		{131072, iobuf.TierLarge},
		{131073, iobuf.TierGreat},
		{524288, iobuf.TierGreat},
		{524289, iobuf.TierHuge},
		{2097152, iobuf.TierHuge},
		{2097153, iobuf.TierVast},
		{8388608, iobuf.TierVast},
		{8388609, iobuf.TierGiant},
		{33554432, iobuf.TierGiant},
		{33554433, iobuf.TierTitan},
		{134217728, iobuf.TierTitan},
		// Larger than max returns Titan
		{134217729, iobuf.TierTitan},
		{1 << 30, iobuf.TierTitan},
	}

	for _, tt := range tests {
		got := iobuf.TierBySize(tt.size)
		if got != tt.expected {
			t.Errorf("TierBySize(%d) = %d, want %d", tt.size, got, tt.expected)
		}
	}
}

func TestBufferTierSize(t *testing.T) {
	tests := []struct {
		tier     iobuf.BufferTier
		expected int
	}{
		{iobuf.TierPico, iobuf.BufferSizePico},
		{iobuf.TierNano, iobuf.BufferSizeNano},
		{iobuf.TierMicro, iobuf.BufferSizeMicro},
		{iobuf.TierSmall, iobuf.BufferSizeSmall},
		{iobuf.TierMedium, iobuf.BufferSizeMedium},
		{iobuf.TierBig, iobuf.BufferSizeBig},
		{iobuf.TierLarge, iobuf.BufferSizeLarge},
		{iobuf.TierGreat, iobuf.BufferSizeGreat},
		{iobuf.TierHuge, iobuf.BufferSizeHuge},
		{iobuf.TierVast, iobuf.BufferSizeVast},
		{iobuf.TierGiant, iobuf.BufferSizeGiant},
		{iobuf.TierTitan, iobuf.BufferSizeTitan},
	}

	for _, tt := range tests {
		got := tt.tier.Size()
		if got != tt.expected {
			t.Errorf("BufferTier(%d).Size() = %d, want %d", tt.tier, got, tt.expected)
		}
	}
}

func TestBufferTierSize_OutOfRange(t *testing.T) {
	// Out of range tiers should return Titan size (max tier)
	if iobuf.BufferTier(-1).Size() != iobuf.BufferSizeTitan {
		t.Errorf("BufferTier(-1).Size() should return BufferSizeTitan")
	}
	if iobuf.TierEnd.Size() != iobuf.BufferSizeTitan {
		t.Errorf("TierEnd.Size() should return BufferSizeTitan")
	}
	if iobuf.BufferTier(100).Size() != iobuf.BufferSizeTitan {
		t.Errorf("BufferTier(100).Size() should return BufferSizeTitan")
	}
}

func TestBufferSizeFor(t *testing.T) {
	tests := []struct {
		size     int
		expected int
	}{
		{1, iobuf.BufferSizePico},
		{32, iobuf.BufferSizePico},
		{33, iobuf.BufferSizeNano},
		{200, iobuf.BufferSizeMicro},
		{1000, iobuf.BufferSizeSmall},
		{8192, iobuf.BufferSizeMedium},
		{10000, iobuf.BufferSizeBig},
		{100000, iobuf.BufferSizeLarge},
		{200000, iobuf.BufferSizeGreat},
		{1000000, iobuf.BufferSizeHuge},
		{5000000, iobuf.BufferSizeVast},
		{20000000, iobuf.BufferSizeGiant},
		{100000000, iobuf.BufferSizeTitan},
		{200000000, iobuf.BufferSizeTitan},
	}

	for _, tt := range tests {
		got := iobuf.BufferSizeFor(tt.size)
		if got != tt.expected {
			t.Errorf("BufferSizeFor(%d) = %d, want %d", tt.size, got, tt.expected)
		}
	}
}

func TestTierConstants(t *testing.T) {
	// Verify tier constants are sequential (12 tiers)
	tiers := []iobuf.BufferTier{
		iobuf.TierPico,
		iobuf.TierNano,
		iobuf.TierMicro,
		iobuf.TierSmall,
		iobuf.TierMedium,
		iobuf.TierBig,
		iobuf.TierLarge,
		iobuf.TierGreat,
		iobuf.TierHuge,
		iobuf.TierVast,
		iobuf.TierGiant,
		iobuf.TierTitan,
	}

	for i, tier := range tiers {
		if int(tier) != i {
			t.Errorf("Tier %d should have value %d, got %d", i, i, tier)
		}
	}

	if int(iobuf.TierEnd) != 12 {
		t.Errorf("TierEnd should be 12, got %d", iobuf.TierEnd)
	}
}

func TestCacheLineSize(t *testing.T) {
	// CacheLineSize should be a positive power of 2 (typically 64 or 128)
	if iobuf.CacheLineSize < 32 || iobuf.CacheLineSize > 256 {
		t.Errorf("CacheLineSize = %d, expected between 32 and 256", iobuf.CacheLineSize)
	}
	// Verify it's a power of 2
	if iobuf.CacheLineSize&(iobuf.CacheLineSize-1) != 0 {
		t.Errorf("CacheLineSize = %d is not a power of 2", iobuf.CacheLineSize)
	}
}

func TestCacheLineAlignedMem(t *testing.T) {
	const size = 1024
	mem := iobuf.CacheLineAlignedMem(size)

	if len(mem) != size {
		t.Errorf("CacheLineAlignedMem length = %d, want %d", len(mem), size)
	}

	ptr := uintptr(unsafe.Pointer(unsafe.SliceData(mem)))
	if ptr%uintptr(iobuf.CacheLineSize) != 0 {
		t.Errorf("CacheLineAlignedMem not cache-line-aligned: address %#x %% %d = %d",
			ptr, iobuf.CacheLineSize, ptr%uintptr(iobuf.CacheLineSize))
	}
}

func TestCacheLineAlignedMem_SmallAllocation(t *testing.T) {
	const size = 16
	mem := iobuf.CacheLineAlignedMem(size)

	if len(mem) != size {
		t.Errorf("CacheLineAlignedMem length = %d, want %d", len(mem), size)
	}

	ptr := uintptr(unsafe.Pointer(unsafe.SliceData(mem)))
	if ptr%uintptr(iobuf.CacheLineSize) != 0 {
		t.Errorf("CacheLineAlignedMem not cache-line-aligned: address %#x", ptr)
	}
}

func TestCacheLineAlignedMemBlocks(t *testing.T) {
	const n = 4
	const blockSize = 128
	blocks := iobuf.CacheLineAlignedMemBlocks(n, blockSize)

	if len(blocks) != n {
		t.Errorf("CacheLineAlignedMemBlocks returned %d blocks, want %d", len(blocks), n)
	}

	for i, block := range blocks {
		if len(block) != blockSize {
			t.Errorf("block[%d] length = %d, want %d", i, len(block), blockSize)
		}
		ptr := uintptr(unsafe.Pointer(unsafe.SliceData(block)))
		if ptr%uintptr(iobuf.CacheLineSize) != 0 {
			t.Errorf("block[%d] not cache-line-aligned: address %#x %% %d = %d",
				i, ptr, iobuf.CacheLineSize, ptr%uintptr(iobuf.CacheLineSize))
		}
	}
}

func TestCacheLineAlignedMemBlocks_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("CacheLineAlignedMemBlocks(0, 64) did not panic")
		}
	}()
	_ = iobuf.CacheLineAlignedMemBlocks(0, 64)
}

func TestSliceOfArray_Panic_Extended(t *testing.T) {
	data := make([]byte, iobuf.BufferSizeTitan*2)

	testCases := []struct {
		name string
		fn   func()
	}{
		{"SliceOfBigArray_n0", func() { iobuf.SliceOfBigArray(data, 0, 0) }},
		{"SliceOfBigArray_nNeg", func() { iobuf.SliceOfBigArray(data, 0, -1) }},
		{"SliceOfGreatArray_n0", func() { iobuf.SliceOfGreatArray(data, 0, 0) }},
		{"SliceOfGreatArray_nNeg", func() { iobuf.SliceOfGreatArray(data, 0, -1) }},
		{"SliceOfVastArray_n0", func() { iobuf.SliceOfVastArray(data, 0, 0) }},
		{"SliceOfVastArray_nNeg", func() { iobuf.SliceOfVastArray(data, 0, -1) }},
		{"SliceOfTitanArray_n0", func() { iobuf.SliceOfTitanArray(data, 0, 0) }},
		{"SliceOfTitanArray_nNeg", func() { iobuf.SliceOfTitanArray(data, 0, -1) }},
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

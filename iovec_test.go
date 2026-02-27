// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package iobuf_test

import (
	"testing"
	"unsafe"

	"code.hybscloud.com/iobuf"
)

const registerBufferSize = iobuf.BufferSizeLarge

func TestIoVecFromBytesSlice(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		vec := iobuf.IoVecFromBytesSlice(nil)
		if vec != nil {
			t.Errorf("expected nil, got %v", vec)
		}
	})

	t.Run("single buffer", func(t *testing.T) {
		buf := make([]byte, 128)
		buf[0] = 0xAB
		iov := [][]byte{buf}
		vec := iobuf.IoVecFromBytesSlice(iov)
		if len(vec) != 1 {
			t.Errorf("expected len=1, got %d", len(vec))
		}
		if vec[0].Base != unsafe.SliceData(buf) {
			t.Error("expected Base to point to buf")
		}
		if vec[0].Len != 128 {
			t.Errorf("expected Len=128, got %d", vec[0].Len)
		}
	})

	t.Run("multiple buffers", func(t *testing.T) {
		bufs := [][]byte{
			make([]byte, 64),
			make([]byte, 128),
			make([]byte, 256),
		}
		vec := iobuf.IoVecFromBytesSlice(bufs)
		if len(vec) != 3 {
			t.Errorf("expected len=3, got %d", len(vec))
		}
		expectedLens := []uint64{64, 128, 256}
		for i, v := range vec {
			if v.Len != expectedLens[i] {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, expectedLens[i])
			}
		}
	})
}

func TestIoVecAddrLen(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		addr, n := iobuf.IoVecAddrLen(nil)
		if addr != 0 || n != 0 {
			t.Errorf("expected (0, 0), got (%d, %d)", addr, n)
		}
	})

	t.Run("non-empty slice", func(t *testing.T) {
		vec := make([]iobuf.IoVec, 4)
		addr, n := iobuf.IoVecAddrLen(vec)
		if n != 4 {
			t.Errorf("expected n=4, got %d", n)
		}
		if addr == 0 {
			t.Error("expected non-zero address")
		}
		expectedAddr := uintptr(unsafe.Pointer(&vec[0]))
		if addr != expectedAddr {
			t.Errorf("expected addr=%d, got %d", expectedAddr, addr)
		}
	})
}

func TestIoVecFrom(t *testing.T) {
	t.Run("PicoBuffer/empty", func(t *testing.T) {
		vec := iobuf.IoVecFrom[iobuf.PicoBuffer](nil)
		if vec != nil {
			t.Error("expected nil for empty input")
		}
	})
	t.Run("PicoBuffer/pointer_and_length", func(t *testing.T) {
		buffers := make([]iobuf.PicoBuffer, 4)
		buffers[0][0] = 0xDE
		buffers[1][0] = 0xAD
		vec := iobuf.IoVecFrom(buffers)
		if len(vec) != 4 {
			t.Errorf("expected len=4, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizePico {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizePico)
			}
			expectedBase := (*byte)(unsafe.Pointer(&buffers[i]))
			if v.Base != expectedBase {
				t.Errorf("vec[%d].Base mismatch", i)
			}
		}
	})

	t.Run("NanoBuffer/empty", func(t *testing.T) {
		if iobuf.IoVecFrom[iobuf.NanoBuffer](nil) != nil {
			t.Error("expected nil for empty input")
		}
	})
	t.Run("NanoBuffer/non-empty", func(t *testing.T) {
		buffers := make([]iobuf.NanoBuffer, 2)
		vec := iobuf.IoVecFrom(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizeNano {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizeNano)
			}
		}
	})

	t.Run("MicroBuffer/empty", func(t *testing.T) {
		if iobuf.IoVecFrom[iobuf.MicroBuffer](nil) != nil {
			t.Error("expected nil for empty input")
		}
	})
	t.Run("MicroBuffer/non-empty", func(t *testing.T) {
		buffers := make([]iobuf.MicroBuffer, 2)
		vec := iobuf.IoVecFrom(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizeMicro {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizeMicro)
			}
		}
	})

	t.Run("SmallBuffer/empty", func(t *testing.T) {
		if iobuf.IoVecFrom[iobuf.SmallBuffer](nil) != nil {
			t.Error("expected nil for empty input")
		}
	})
	t.Run("SmallBuffer/non-empty", func(t *testing.T) {
		buffers := make([]iobuf.SmallBuffer, 2)
		vec := iobuf.IoVecFrom(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizeSmall {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizeSmall)
			}
		}
	})

	t.Run("MediumBuffer/empty", func(t *testing.T) {
		if iobuf.IoVecFrom[iobuf.MediumBuffer](nil) != nil {
			t.Error("expected nil for empty input")
		}
	})
	t.Run("MediumBuffer/non-empty", func(t *testing.T) {
		buffers := make([]iobuf.MediumBuffer, 2)
		vec := iobuf.IoVecFrom(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizeMedium {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizeMedium)
			}
		}
	})

	t.Run("BigBuffer/empty", func(t *testing.T) {
		if iobuf.IoVecFrom[iobuf.BigBuffer](nil) != nil {
			t.Error("expected nil for empty input")
		}
	})
	t.Run("BigBuffer/non-empty", func(t *testing.T) {
		buffers := make([]iobuf.BigBuffer, 2)
		vec := iobuf.IoVecFrom(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizeBig {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizeBig)
			}
		}
	})

	t.Run("LargeBuffer/empty", func(t *testing.T) {
		if iobuf.IoVecFrom[iobuf.LargeBuffer](nil) != nil {
			t.Error("expected nil for empty input")
		}
	})
	t.Run("LargeBuffer/non-empty", func(t *testing.T) {
		buffers := make([]iobuf.LargeBuffer, 2)
		vec := iobuf.IoVecFrom(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizeLarge {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizeLarge)
			}
		}
	})

	t.Run("GreatBuffer/empty", func(t *testing.T) {
		if iobuf.IoVecFrom[iobuf.GreatBuffer](nil) != nil {
			t.Error("expected nil for empty input")
		}
	})
	t.Run("GreatBuffer/non-empty", func(t *testing.T) {
		buffers := make([]iobuf.GreatBuffer, 2)
		vec := iobuf.IoVecFrom(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizeGreat {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizeGreat)
			}
		}
	})

	t.Run("HugeBuffer/empty", func(t *testing.T) {
		if iobuf.IoVecFrom[iobuf.HugeBuffer](nil) != nil {
			t.Error("expected nil for empty input")
		}
	})
	t.Run("HugeBuffer/non-empty", func(t *testing.T) {
		buffers := make([]iobuf.HugeBuffer, 2)
		vec := iobuf.IoVecFrom(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizeHuge {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizeHuge)
			}
		}
	})

	t.Run("VastBuffer/empty", func(t *testing.T) {
		if iobuf.IoVecFrom[iobuf.VastBuffer](nil) != nil {
			t.Error("expected nil for empty input")
		}
	})
	t.Run("VastBuffer/non-empty", func(t *testing.T) {
		buffers := make([]iobuf.VastBuffer, 2)
		vec := iobuf.IoVecFrom(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizeVast {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizeVast)
			}
		}
	})

	t.Run("GiantBuffer/empty", func(t *testing.T) {
		if iobuf.IoVecFrom[iobuf.GiantBuffer](nil) != nil {
			t.Error("expected nil for empty input")
		}
	})
	t.Run("GiantBuffer/non-empty", func(t *testing.T) {
		buffers := make([]iobuf.GiantBuffer, 2)
		vec := iobuf.IoVecFrom(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizeGiant {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizeGiant)
			}
		}
	})

	t.Run("TitanBuffer/empty", func(t *testing.T) {
		if iobuf.IoVecFrom[iobuf.TitanBuffer](nil) != nil {
			t.Error("expected nil for empty input")
		}
	})
	t.Run("TitanBuffer/non-empty", func(t *testing.T) {
		buffers := make([]iobuf.TitanBuffer, 1)
		vec := iobuf.IoVecFrom(buffers)
		if len(vec) != 1 {
			t.Errorf("expected len=1, got %d", len(vec))
		}
		if vec[0].Len != iobuf.BufferSizeTitan {
			t.Errorf("vec[0].Len = %d, expected %d", vec[0].Len, iobuf.BufferSizeTitan)
		}
	})
}

func TestIoVecFromRegisteredBuffers(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		vec := iobuf.IoVecFromRegisteredBuffers(nil)
		if vec != nil {
			t.Error("expected nil for empty input")
		}
	})

	t.Run("pointer and length correctness", func(t *testing.T) {
		buffers := make([]iobuf.RegisterBuffer, 2)
		vec := iobuf.IoVecFromRegisteredBuffers(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != registerBufferSize {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, registerBufferSize)
			}
			expectedBase := (*byte)(unsafe.Pointer(&buffers[i]))
			if v.Base != expectedBase {
				t.Errorf("vec[%d].Base mismatch", i)
			}
		}
	})
}

func TestIoVecPointerStability(t *testing.T) {
	buffers := make([]iobuf.PicoBuffer, 4)
	buffers[0][0] = 0x11
	buffers[1][0] = 0x22
	buffers[2][0] = 0x33
	buffers[3][0] = 0x44

	vec := iobuf.IoVecFrom(buffers)

	for i := range vec {
		ptr := unsafe.Pointer(vec[i].Base)
		val := *(*byte)(ptr)
		expected := byte((i + 1) * 0x11)
		if val != expected {
			t.Errorf("vec[%d] points to value 0x%02X, expected 0x%02X", i, val, expected)
		}
	}
}

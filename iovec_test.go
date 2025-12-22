// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package iobuf_test

import (
	"testing"
	"unsafe"

	"code.hybscloud.com/iobuf"
)

const registerBufferSize = iobuf.BufferSizeHuge

func TestIoVecFromBytesSlice(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		addr, n := iobuf.IoVecFromBytesSlice(nil)
		if addr != 0 || n != 0 {
			t.Errorf("expected (0, 0), got (%d, %d)", addr, n)
		}
	})

	t.Run("single buffer", func(t *testing.T) {
		buf := make([]byte, 128)
		buf[0] = 0xAB
		iov := [][]byte{buf}
		addr, n := iobuf.IoVecFromBytesSlice(iov)
		if n != 1 {
			t.Errorf("expected n=1, got %d", n)
		}
		if addr == 0 {
			t.Error("expected non-zero address")
		}
	})

	t.Run("multiple buffers", func(t *testing.T) {
		bufs := [][]byte{
			make([]byte, 64),
			make([]byte, 128),
			make([]byte, 256),
		}
		addr, n := iobuf.IoVecFromBytesSlice(bufs)
		if n != 3 {
			t.Errorf("expected n=3, got %d", n)
		}
		if addr == 0 {
			t.Error("expected non-zero address")
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

func TestIoVecFromPicoBuffers(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		vec := iobuf.IoVecFromPicoBuffers(nil)
		if vec != nil {
			t.Error("expected nil for empty input")
		}
	})

	t.Run("pointer and length correctness", func(t *testing.T) {
		buffers := make([]iobuf.PicoBuffer, 4)
		buffers[0][0] = 0xDE
		buffers[1][0] = 0xAD
		vec := iobuf.IoVecFromPicoBuffers(buffers)
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
}

func TestIoVecFromNanoBuffers(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		vec := iobuf.IoVecFromNanoBuffers(nil)
		if vec != nil {
			t.Error("expected nil for empty input")
		}
	})

	t.Run("non-empty", func(t *testing.T) {
		buffers := make([]iobuf.NanoBuffer, 2)
		vec := iobuf.IoVecFromNanoBuffers(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizeNano {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizeNano)
			}
		}
	})
}

func TestIoVecFromMicroBuffers(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		vec := iobuf.IoVecFromMicroBuffers(nil)
		if vec != nil {
			t.Error("expected nil for empty input")
		}
	})

	t.Run("non-empty", func(t *testing.T) {
		buffers := make([]iobuf.MicroBuffer, 2)
		vec := iobuf.IoVecFromMicroBuffers(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizeMicro {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizeMicro)
			}
		}
	})
}

func TestIoVecFromSmallBuffers(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		vec := iobuf.IoVecFromSmallBuffers(nil)
		if vec != nil {
			t.Error("expected nil for empty input")
		}
	})

	t.Run("non-empty", func(t *testing.T) {
		buffers := make([]iobuf.SmallBuffer, 2)
		vec := iobuf.IoVecFromSmallBuffers(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizeSmall {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizeSmall)
			}
		}
	})
}

func TestIoVecFromMediumBuffers(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		vec := iobuf.IoVecFromMediumBuffers(nil)
		if vec != nil {
			t.Error("expected nil for empty input")
		}
	})

	t.Run("non-empty", func(t *testing.T) {
		buffers := make([]iobuf.MediumBuffer, 2)
		vec := iobuf.IoVecFromMediumBuffers(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizeMedium {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizeMedium)
			}
		}
	})
}

func TestIoVecFromLargeBuffers(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		vec := iobuf.IoVecFromLargeBuffers(nil)
		if vec != nil {
			t.Error("expected nil for empty input")
		}
	})

	t.Run("non-empty", func(t *testing.T) {
		buffers := make([]iobuf.LargeBuffer, 2)
		vec := iobuf.IoVecFromLargeBuffers(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizeLarge {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizeLarge)
			}
		}
	})
}

func TestIoVecFromHugeBuffers(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		vec := iobuf.IoVecFromHugeBuffers(nil)
		if vec != nil {
			t.Error("expected nil for empty input")
		}
	})

	t.Run("non-empty", func(t *testing.T) {
		buffers := make([]iobuf.HugeBuffer, 2)
		vec := iobuf.IoVecFromHugeBuffers(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizeHuge {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizeHuge)
			}
		}
	})
}

func TestIoVecFromGiantBuffers(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		vec := iobuf.IoVecFromGiantBuffers(nil)
		if vec != nil {
			t.Error("expected nil for empty input")
		}
	})

	t.Run("non-empty", func(t *testing.T) {
		buffers := make([]iobuf.GiantBuffer, 2)
		vec := iobuf.IoVecFromGiantBuffers(buffers)
		if len(vec) != 2 {
			t.Errorf("expected len=2, got %d", len(vec))
		}
		for i, v := range vec {
			if v.Len != iobuf.BufferSizeGiant {
				t.Errorf("vec[%d].Len = %d, expected %d", i, v.Len, iobuf.BufferSizeGiant)
			}
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

	vec := iobuf.IoVecFromPicoBuffers(buffers)

	for i := range vec {
		ptr := unsafe.Pointer(vec[i].Base)
		val := *(*byte)(ptr)
		expected := byte((i + 1) * 0x11)
		if val != expected {
			t.Errorf("vec[%d] points to value 0x%02X, expected 0x%02X", i, val, expected)
		}
	}
}

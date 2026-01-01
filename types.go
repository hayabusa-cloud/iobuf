// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package iobuf

import "net"

// PageSize is the memory page size used for aligned allocations.
//
// The default value (4 KiB) matches the typical x86-64 and ARM64 page size.
// Use SetPageSize to configure for systems with different page sizes.
var PageSize uintptr = 4096

// SetPageSize updates the package-level page size used for aligned allocations.
//
// This should be called once during initialization, before any calls to
// AlignedMem or AlignedMemBlocks. Common values:
//   - 4096 (4 KiB): Standard x86-64, ARM64
//   - 16384 (16 KiB): Some ARM64 configurations (Apple Silicon)
//   - 65536 (64 KiB): Some embedded systems
func SetPageSize(size int) {
	PageSize = uintptr(size)
}

// Buffers is an alias for net.Buffers, providing a standard way to group
// multiple byte slices for vectored I/O operations.
type Buffers = net.Buffers

// noCopy is a sentinel type that triggers "go vet" warnings when a
// containing struct is copied by value.
//
// Embedding this type in a struct (e.g., BoundedPool) causes go vet to
// report "copies lock value" when the struct is passed by value or assigned.
// This is a compile-time safety mechanism for types that must not be copied.
//
// The Lock/Unlock methods satisfy the sync.Locker interface, which is
// the detection mechanism used by go vet's copylock analyzer.
type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

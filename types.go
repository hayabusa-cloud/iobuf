// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package iobuf

import "net"

// PageSize defines the standard memory page size (4 KiB) used for alignment.
var PageSize uintptr = 4096

// SetPageSize updates the package-level page size used for allocations.
func SetPageSize(size int) {
	PageSize = uintptr(size)
}

// Buffers is an alias for net.Buffers, providing a standard way to group
// multiple byte slices for vectored I/O operations.
type Buffers = net.Buffers

// noCopy is a sentinel used to prevent copying of synchronization primitives.
type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

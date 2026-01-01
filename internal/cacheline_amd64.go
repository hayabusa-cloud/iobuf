// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build amd64

package internal

// CacheLineSize is the L1 cache line size for x86-64 architectures.
// All modern Intel and AMD processors use 64-byte cache lines.
const CacheLineSize = 64

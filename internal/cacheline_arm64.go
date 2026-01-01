// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build arm64

package internal

// CacheLineSize is the L1 cache line size for ARM64 architectures.
// Apple Silicon (M1/M2/M3) uses 128-byte L2 cache lines, but L1 is 64 bytes.
// Most ARM Cortex-A series use 64-byte L1 cache lines.
// Use 128 bytes as conservative value for Apple Silicon compatibility.
const CacheLineSize = 128

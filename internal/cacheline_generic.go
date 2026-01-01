// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build !amd64 && !arm64 && !riscv64 && !loong64

package internal

// CacheLineSize is the default L1 cache line size for other 64-bit architectures.
// 64 bytes is the most common cache line size on modern CPUs.
// Covers: mips64, mips64le, ppc64, ppc64le, s390x, wasm, sparc64, etc.
//
// Note: 32-bit architectures (386, arm, mips, mipsle) are not supported
// by this module due to 64-bit integer requirements in BoundedPool.
const CacheLineSize = 64

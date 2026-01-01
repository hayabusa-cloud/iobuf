// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build riscv64

package internal

// CacheLineSize is the L1 cache line size for RISC-V 64-bit architectures.
// Common implementations (SiFive, T-Head) use 64-byte cache lines.
const CacheLineSize = 64

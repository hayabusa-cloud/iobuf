// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build 386 || arm || mips || mipsle || ppc || s390 || armbe || mipsbe || riscv32

package internal

// CacheLineSize is the typical SIMD/L1 cache line size for 32-bit architectures.
const CacheLineSize = 32

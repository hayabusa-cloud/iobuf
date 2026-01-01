// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build loong64

package internal

// CacheLineSize is the L1 cache line size for LoongArch 64-bit architectures.
// Loongson 3A5000/3A6000 series use 64-byte cache lines.
const CacheLineSize = 64

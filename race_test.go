// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build race

package iobuf_test

// raceEnabled is true when the race detector is active.
// TitanBuffer tests are skipped in race mode due to stack overhead.
const raceEnabled = true

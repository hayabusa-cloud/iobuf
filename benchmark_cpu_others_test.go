// ©Hayabusa Cloud Co., Ltd. 2026. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build !unix

package iobuf_test

import (
	"testing"
	"time"
)

func benchmarkCPUTime(*testing.B) time.Duration { return 0 }

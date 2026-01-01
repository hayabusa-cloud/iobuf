// ©Hayabusa Cloud Co., Ltd. 2025. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package iobuf

import "testing"

// TestNoCopy tests the noCopy sentinel type.
// noCopy implements sync.Locker interface for go vet copy detection.
func TestNoCopy(t *testing.T) {
	var nc noCopy
	nc.Lock()
	nc.Unlock()
}

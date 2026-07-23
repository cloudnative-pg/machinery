//go:build linux || darwin

/*
Copyright © contributors to CloudNativePG, established as
CloudNativePG a Series of LF Projects, LLC.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

SPDX-License-Identifier: Apache-2.0
*/

package log

import (
	"io"
	"path/filepath"

	"golang.org/x/sys/unix"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These specs pin the "(ignored for FIFOs)" promise made by the
// --log-truncate-destination flag and SetTruncateDestination: opening a FIFO
// destination must succeed regardless of the truncate setting, since a FIFO
// has no byte offset for O_APPEND or O_TRUNC to act on.
var _ = Describe("customDestination with a FIFO destination", func() {
	AfterEach(func() {
		logDestination = ""
		truncateDestination = false
	})

	openFifoDestination := func() {
		fifoPath := filepath.Join(GinkgoT().TempDir(), "log.fifo")
		Expect(unix.Mkfifo(fifoPath, 0o600)).To(Succeed())

		logDestination = fifoPath
		options := &zap.Options{}
		// O_RDWR keeps the open from blocking on a missing peer, so this
		// returns without a reader present.
		Expect(func() { customDestination(options) }).NotTo(Panic())
		Expect(options.DestWriter).NotTo(BeNil())

		closer, ok := options.DestWriter.(io.Closer)
		Expect(ok).To(BeTrue())
		Expect(closer.Close()).To(Succeed())
	}

	It("opens the FIFO without error in append (default) mode", func() {
		openFifoDestination()
	})

	It("opens the FIFO without error in truncate mode", func() {
		SetTruncateDestination(true)
		openFifoDestination()
	})
})

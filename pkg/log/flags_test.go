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
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("customDestination", func() {
	AfterEach(func() {
		logDestination = ""
		truncateDestination = false
	})

	It("does nothing when no log destination is configured", func() {
		options := &zap.Options{}
		customDestination(options)
		Expect(options.DestWriter).To(BeNil())
	})

	It("appends to an existing regular file instead of overwriting from offset 0", func() {
		destPath := filepath.Join(GinkgoT().TempDir(), "log.txt")
		Expect(os.WriteFile(destPath, []byte("0123456789"), 0o600)).To(Succeed())

		logDestination = destPath
		options := &zap.Options{}
		customDestination(options)
		Expect(options.DestWriter).NotTo(BeNil())

		_, err := options.DestWriter.Write([]byte("abc"))
		Expect(err).NotTo(HaveOccurred())

		if closer, ok := options.DestWriter.(io.Closer); ok {
			Expect(closer.Close()).To(Succeed())
		}

		content, err := os.ReadFile(destPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("0123456789abc"))
	})

	It("keeps appending across repeated opens, as would happen across separate process invocations", func() {
		destPath := filepath.Join(GinkgoT().TempDir(), "log.txt")
		logDestination = destPath

		firstOpen := &zap.Options{}
		customDestination(firstOpen)
		_, err := firstOpen.DestWriter.Write([]byte("first-run\n"))
		Expect(err).NotTo(HaveOccurred())
		Expect(firstOpen.DestWriter.(io.Closer).Close()).To(Succeed())

		secondOpen := &zap.Options{}
		customDestination(secondOpen)
		_, err = secondOpen.DestWriter.Write([]byte("second\n"))
		Expect(err).NotTo(HaveOccurred())
		Expect(secondOpen.DestWriter.(io.Closer).Close()).To(Succeed())

		content, err := os.ReadFile(destPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("first-run\nsecond\n"))
	})

	It("truncates the destination when SetTruncateDestination(true) was called", func() {
		destPath := filepath.Join(GinkgoT().TempDir(), "log.txt")
		Expect(os.WriteFile(destPath, []byte("0123456789"), 0o600)).To(Succeed())

		SetTruncateDestination(true)
		defer SetTruncateDestination(false)

		logDestination = destPath
		options := &zap.Options{}
		customDestination(options)
		Expect(options.DestWriter).NotTo(BeNil())

		_, err := options.DestWriter.Write([]byte("abc"))
		Expect(err).NotTo(HaveOccurred())
		Expect(options.DestWriter.(io.Closer).Close()).To(Succeed())

		content, err := os.ReadFile(destPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("abc"))
	})
})

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

package compatibility

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateFifo", func() {
	It("creates a FIFO when nothing exists at the path yet", func() {
		fifoPath := filepath.Join(GinkgoT().TempDir(), "myfifo")

		Expect(CreateFifo(fifoPath)).To(Succeed())

		info, err := os.Lstat(fifoPath) //#nosec
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode() & os.ModeNamedPipe).NotTo(BeZero())
	})

	It("is a no-op when a FIFO already exists at the path", func() {
		fifoPath := filepath.Join(GinkgoT().TempDir(), "myfifo")
		Expect(CreateFifo(fifoPath)).To(Succeed())

		// attempt to create a FIFO where one already exists
		Expect(CreateFifo(fifoPath)).To(Succeed())

		info, err := os.Lstat(fifoPath) //#nosec
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode() & os.ModeNamedPipe).NotTo(BeZero())
	})

	It("fails loud when a regular file already exists at the path", func() {
		filePath := filepath.Join(GinkgoT().TempDir(), "notafifo")
		Expect(os.WriteFile(filePath, []byte("hello"), 0o600)).To(Succeed())

		// attempt to create a FIFO where a regular file already exists
		err := CreateFifo(filePath)
		Expect(err).To(MatchError(ErrExistsNotFifo))
		Expect(err.Error()).To(ContainSubstring(filePath))

		// the pre-existing file must be left untouched
		content, readErr := os.ReadFile(filePath) //#nosec
		Expect(readErr).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("hello"))
	})

	It("propagates Stat errors other than not-exist instead of attempting Mkfifo", func() {
		filePath := filepath.Join(GinkgoT().TempDir(), "notadir")
		Expect(os.WriteFile(filePath, []byte("hello"), 0o600)).To(Succeed())

		// attempt to create a FIFO in a non-directory, which will cause Stat to fail with an error other than "not exist"
		err := CreateFifo(filepath.Join(filePath, "myfifo"))
		Expect(err).To(HaveOccurred())
		Expect(os.IsNotExist(err)).To(BeFalse())
	})

	It("accepts a symlink that resolves to a FIFO, since Stat follows it like the reader does", func() {
		dir := GinkgoT().TempDir()
		realFifo := filepath.Join(dir, "real.fifo")
		Expect(CreateFifo(realFifo)).To(Succeed())

		// A symlink pointing at a genuine FIFO resolves to a FIFO under
		// os.Stat, matching how consumers of this path (os.OpenFile) open
		// it: they follow the link too, so the check must agree with them.
		linkPath := filepath.Join(dir, "link.fifo")
		Expect(os.Symlink(realFifo, linkPath)).To(Succeed())

		Expect(CreateFifo(linkPath)).To(Succeed())
	})

	It("rejects a symlink that resolves to a regular file", func() {
		dir := GinkgoT().TempDir()
		target := filepath.Join(dir, "notafifo")
		Expect(os.WriteFile(target, []byte("hello"), 0o600)).To(Succeed())

		linkPath := filepath.Join(dir, "link")
		Expect(os.Symlink(target, linkPath)).To(Succeed())

		err := CreateFifo(linkPath)
		Expect(err).To(MatchError(ErrExistsNotFifo))

		// the pre-existing target must be left untouched
		content, readErr := os.ReadFile(target) //#nosec
		Expect(readErr).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("hello"))
	})
})

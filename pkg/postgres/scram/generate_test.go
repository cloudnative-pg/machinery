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

package scram

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GenerateOptions Defaults", func() {
	It("fills in default iterations when zero", func() {
		opts := &GenerateOptions{PlainText: "secret"}
		Expect(opts.Defaults()).To(Succeed())
		Expect(opts.Iterations).To(Equal(DefaultPostgresIterations))
	})

	It("preserves a custom iteration count", func() {
		opts := &GenerateOptions{PlainText: "secret", Iterations: 1024}
		Expect(opts.Defaults()).To(Succeed())
		Expect(opts.Iterations).To(Equal(1024))
	})

	It("generates a salt of the expected length when missing", func() {
		opts := &GenerateOptions{PlainText: "secret"}
		Expect(opts.Defaults()).To(Succeed())
		Expect(opts.Salt).To(HaveLen(DefaultSaltLength))
	})

	It("preserves a user-provided salt", func() {
		salt := []byte("0123456789abcdef")
		opts := &GenerateOptions{PlainText: "secret", Salt: salt}
		Expect(opts.Defaults()).To(Succeed())
		Expect(opts.Salt).To(Equal(salt))
	})
})

var _ = Describe("GenerateOptions Generate", func() {
	It("produces a hash in the PostgreSQL SCRAM-SHA-256 format", func() {
		opts := &GenerateOptions{
			PlainText:  "secret",
			Salt:       []byte("0123456789abcdef"),
			Iterations: 4096,
		}

		hash, err := opts.Generate()
		Expect(err).ToNot(HaveOccurred())
		Expect(strings.HasPrefix(hash, "SCRAM-SHA-256$4096:")).To(BeTrue())

		parts := strings.Split(hash, "$")
		Expect(parts).To(HaveLen(3))
		Expect(strings.Split(parts[2], ":")).To(HaveLen(2))
	})

	It("populates defaults when called without explicit options", func() {
		opts := &GenerateOptions{PlainText: "secret"}
		hash, err := opts.Generate()
		Expect(err).ToNot(HaveOccurred())
		Expect(strings.HasPrefix(hash, "SCRAM-SHA-256$4096:")).To(BeTrue())
		Expect(opts.Salt).To(HaveLen(DefaultSaltLength))
	})

	It("is deterministic for the same salt, iterations and password", func() {
		opts := &GenerateOptions{
			PlainText:  "secret",
			Salt:       []byte("0123456789abcdef"),
			Iterations: 4096,
		}
		first, err := opts.Generate()
		Expect(err).ToNot(HaveOccurred())
		second, err := opts.Generate()
		Expect(err).ToNot(HaveOccurred())
		Expect(first).To(Equal(second))
	})

	It("produces different hashes when the salt changes", func() {
		opts1 := &GenerateOptions{
			PlainText:  "secret",
			Salt:       []byte("aaaaaaaaaaaaaaaa"),
			Iterations: 4096,
		}
		opts2 := &GenerateOptions{
			PlainText:  "secret",
			Salt:       []byte("bbbbbbbbbbbbbbbb"),
			Iterations: 4096,
		}
		h1, err := opts1.Generate()
		Expect(err).ToNot(HaveOccurred())
		h2, err := opts2.Generate()
		Expect(err).ToNot(HaveOccurred())
		Expect(h1).ToNot(Equal(h2))
	})
})

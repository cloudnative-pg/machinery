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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Verify", func() {
	const plainText = "secret"

	var hash string

	BeforeEach(func() {
		opts := &GenerateOptions{
			PlainText:  plainText,
			Salt:       []byte("0123456789abcdef"),
			Iterations: 4096,
		}
		var err error
		hash, err = opts.Generate()
		Expect(err).ToNot(HaveOccurred())
	})

	It("returns true for the correct password", func() {
		ok, err := Verify(hash, plainText)
		Expect(err).ToNot(HaveOccurred())
		Expect(ok).To(BeTrue())
	})

	It("returns false for a wrong password", func() {
		ok, err := Verify(hash, "wrong-password")
		Expect(err).ToNot(HaveOccurred())
		Expect(ok).To(BeFalse())
	})
})

var _ = Describe("parsePostgreSQLHash", func() {
	It("parses a well-formed hash", func() {
		opts := &GenerateOptions{
			PlainText:  "secret",
			Salt:       []byte("0123456789abcdef"),
			Iterations: 4096,
		}
		hash, err := opts.Generate()
		Expect(err).ToNot(HaveOccurred())

		parsed, err := parsePostgreSQLHash(hash)
		Expect(err).ToNot(HaveOccurred())
		Expect(parsed.Iterations).To(Equal(4096))
		Expect(string(parsed.RawSalt)).To(Equal("0123456789abcdef"))
		Expect(parsed.RawStoredKey).ToNot(BeEmpty())
		Expect(parsed.RawServerKey).ToNot(BeEmpty())
	})

	It("rejects a hash with the wrong number of '$' sections", func() {
		_, err := parsePostgreSQLHash("SCRAM-SHA-256$4096:salt")
		Expect(err).To(MatchError(ErrWrongComponents))
	})

	It("rejects an unsupported hash type", func() {
		_, err := parsePostgreSQLHash("MD5$4096:c2FsdA==$c3RvcmVk:c2VydmVy")
		Expect(err).To(MatchError(ErrWrongHashType))
	})

	It("rejects a malformed iteration/salt block", func() {
		_, err := parsePostgreSQLHash("SCRAM-SHA-256$4096$c3RvcmVk:c2VydmVy")
		Expect(err).To(MatchError(ErrWrongHashConfig))
	})

	It("rejects a malformed key block", func() {
		_, err := parsePostgreSQLHash("SCRAM-SHA-256$4096:c2FsdA==$c3RvcmVkc2VydmVy")
		Expect(err).To(MatchError(ErrWrongKeyComponents))
	})

	It("rejects a non-numeric iteration count", func() {
		_, err := parsePostgreSQLHash("SCRAM-SHA-256$abc:c2FsdA==$c3RvcmVk:c2VydmVy")
		Expect(err).To(HaveOccurred())
	})

	It("rejects a non-base64 salt", func() {
		_, err := parsePostgreSQLHash("SCRAM-SHA-256$4096:!!!$c3RvcmVk:c2VydmVy")
		Expect(err).To(HaveOccurred())
	})

	It("rejects a non-base64 stored key", func() {
		_, err := parsePostgreSQLHash("SCRAM-SHA-256$4096:c2FsdA==$!!!:c2VydmVy")
		Expect(err).To(HaveOccurred())
	})

	It("rejects a non-base64 server key", func() {
		// Use a properly-sized StoredKey so the failure is the ServerKey's
		// base64 decoding and not the StoredKey's length check.
		_, err := parsePostgreSQLHash(
			"SCRAM-SHA-256$4096:c2FsdA==$bpSY5Ze9NUH+I35LC3gVq+DpBfK46iXBxvhAKqVu9pE=:!!!")
		Expect(err).To(HaveOccurred())
	})

	It("rejects a zero iteration count", func() {
		_, err := parsePostgreSQLHash("SCRAM-SHA-256$0:c2FsdA==$" +
			"bpSY5Ze9NUH+I35LC3gVq+DpBfK46iXBxvhAKqVu9pE=:" +
			"VpYlBuxyzeCI1KnctrefdljpB1mk3Gp7sBI/t11+NkQ=")
		Expect(err).To(MatchError(ErrInvalidIterations))
	})

	It("rejects a negative iteration count", func() {
		_, err := parsePostgreSQLHash("SCRAM-SHA-256$-1:c2FsdA==$" +
			"bpSY5Ze9NUH+I35LC3gVq+DpBfK46iXBxvhAKqVu9pE=:" +
			"VpYlBuxyzeCI1KnctrefdljpB1mk3Gp7sBI/t11+NkQ=")
		Expect(err).To(MatchError(ErrInvalidIterations))
	})

	It("rejects a stored key of wrong length", func() {
		_, err := parsePostgreSQLHash("SCRAM-SHA-256$4096:c2FsdA==$c3RvcmVk:" +
			"VpYlBuxyzeCI1KnctrefdljpB1mk3Gp7sBI/t11+NkQ=")
		Expect(err).To(MatchError(ErrInvalidStoredKey))
	})

	It("rejects a server key of wrong length", func() {
		_, err := parsePostgreSQLHash("SCRAM-SHA-256$4096:c2FsdA==$" +
			"bpSY5Ze9NUH+I35LC3gVq+DpBfK46iXBxvhAKqVu9pE=:c2VydmVy")
		Expect(err).To(MatchError(ErrInvalidServerKey))
	})
})

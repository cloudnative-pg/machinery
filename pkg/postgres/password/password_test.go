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

package password

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// validSCRAM was generated with pkg/postgres/scram using the deterministic
// inputs PlainText="secret", Salt=[]byte("0123456789abcdef"), Iterations=4096.
const validSCRAM = "SCRAM-SHA-256$4096:MDEyMzQ1Njc4OWFiY2RlZg==$" +
	"bpSY5Ze9NUH+I35LC3gVq+DpBfK46iXBxvhAKqVu9pE=:" +
	"VpYlBuxyzeCI1KnctrefdljpB1mk3Gp7sBI/t11+NkQ="

var _ = Describe("GetType", func() {
	DescribeTable("classifies known shadow_pass shapes",
		func(in string, expected Type) {
			Expect(GetType(in)).To(Equal(expected))
		},

		// MD5
		Entry("a valid MD5 hash", "md5"+"0123456789abcdef0123456789abcdef", MD5),
		Entry("an MD5 hash with all-f digits", "md5"+"ffffffffffffffffffffffffffffffff", MD5),

		// SCRAM-SHA-256
		Entry("a valid SCRAM-SHA-256 secret", validSCRAM, SCRAMSHA256),

		// Plaintext / unrecognized
		Entry("empty string", "", Plaintext),
		Entry("an arbitrary plaintext password", "hunter2", Plaintext),
		Entry("MD5 prefix with uppercase hex", "md5"+"ABCDEF0123456789ABCDEF0123456789", Plaintext),
		Entry("MD5 prefix with wrong length", "md5"+"abc", Plaintext),
		Entry("MD5 prefix with non-hex character", "md5"+"0123456789abcdef0123456789abcdez", Plaintext),
		Entry("SCRAM prefix but only two sections", "SCRAM-SHA-256$4096:c2FsdA==", Plaintext),
		Entry("SCRAM with unsupported hash type", "SCRAM-SHA-512$4096:c2FsdA==$"+
			"bpSY5Ze9NUH+I35LC3gVq+DpBfK46iXBxvhAKqVu9pE=:"+
			"VpYlBuxyzeCI1KnctrefdljpB1mk3Gp7sBI/t11+NkQ=", Plaintext),
		Entry("SCRAM with non-numeric iterations", "SCRAM-SHA-256$abc:c2FsdA==$"+
			"bpSY5Ze9NUH+I35LC3gVq+DpBfK46iXBxvhAKqVu9pE=:"+
			"VpYlBuxyzeCI1KnctrefdljpB1mk3Gp7sBI/t11+NkQ=", Plaintext),
		Entry("SCRAM with zero iterations", "SCRAM-SHA-256$0:c2FsdA==$"+
			"bpSY5Ze9NUH+I35LC3gVq+DpBfK46iXBxvhAKqVu9pE=:"+
			"VpYlBuxyzeCI1KnctrefdljpB1mk3Gp7sBI/t11+NkQ=", Plaintext),
		Entry("SCRAM with non-base64 salt", "SCRAM-SHA-256$4096:!!!$"+
			"bpSY5Ze9NUH+I35LC3gVq+DpBfK46iXBxvhAKqVu9pE=:"+
			"VpYlBuxyzeCI1KnctrefdljpB1mk3Gp7sBI/t11+NkQ=", Plaintext),
		Entry("SCRAM with stored key of wrong length",
			"SCRAM-SHA-256$4096:c2FsdA==$c3RvcmVk:"+
				"VpYlBuxyzeCI1KnctrefdljpB1mk3Gp7sBI/t11+NkQ=", Plaintext),
		Entry("SCRAM with server key of wrong length",
			"SCRAM-SHA-256$4096:c2FsdA==$"+
				"bpSY5Ze9NUH+I35LC3gVq+DpBfK46iXBxvhAKqVu9pE=:c2VydmVy", Plaintext),
	)
})

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
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// makeSalt creates a random slice of bytes of DefaultSaltLength
func makeSalt() ([]byte, error) {
	salt := make([]byte, DefaultSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// formatHash renders a PostgreSQL SCRAM-SHA-256 password hash from its
// components in the canonical "SCRAM-SHA-256$<iter>:<salt>$<StoredKey>:<ServerKey>"
// form.
func formatHash(rawSalt []byte, iterations int, storedKey, serverKey []byte) string {
	return fmt.Sprintf("SCRAM-SHA-256$%d:%s$%s:%s",
		iterations,
		base64.StdEncoding.EncodeToString(rawSalt),
		base64.StdEncoding.EncodeToString(storedKey),
		base64.StdEncoding.EncodeToString(serverKey),
	)
}

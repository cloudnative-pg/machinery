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

// Package password classifies the value stored in pg_authid.rolpassword
// the same way PostgreSQL does, so that callers can decide whether a
// proposed password is already encoded or still in plaintext.
package password

import (
	"encoding/base64"
	"strconv"
	"strings"
)

// Type classifies a shadow password value, matching PostgreSQL's PasswordType
// enum from src/include/libpq/crypt.h.
type Type int

const (
	// Plaintext means the value is not a recognized password hash.
	Plaintext Type = iota
	// MD5 means the value is an MD5-hashed password ("md5" + 32 hex digits).
	MD5
	// SCRAMSHA256 means the value is a SCRAM-SHA-256 secret in PostgreSQL's
	// canonical "SCRAM-SHA-256$<iter>:<salt>$<StoredKey>:<ServerKey>" form.
	SCRAMSHA256
)

// Constants mirrored from PostgreSQL's source.
//
// See:
//   - src/include/common/md5.h        (MD5_PASSWD_LEN, MD5_PASSWD_CHARSET)
//   - src/include/common/scram-common.h (SCRAM_SHA_256_KEY_LEN)
const (
	md5PasswordLen    = 35 // "md5" prefix plus 32 hex digits
	scramSHA256KeyLen = 32 // SHA-256 digest size
)

// GetType reports how PostgreSQL would classify the given shadow_pass value.
//
// This is the Go equivalent of get_password_type in PostgreSQL's
// src/backend/libpq/crypt.c. The structural SCRAM check mirrors
// parse_scram_secret in src/backend/libpq/auth-scram.c. Keep these in sync
// with upstream when the supported PostgreSQL versions change.
func GetType(shadowPass string) Type {
	if isMD5(shadowPass) {
		return MD5
	}
	if isSCRAMSHA256(shadowPass) {
		return SCRAMSHA256
	}
	return Plaintext
}

// isMD5 reports whether s matches PostgreSQL's MD5 password format:
// the literal prefix "md5" followed by exactly 32 lowercase hex digits.
func isMD5(s string) bool {
	if len(s) != md5PasswordLen {
		return false
	}
	if !strings.HasPrefix(s, "md5") {
		return false
	}
	for _, c := range s[3:] {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// isSCRAMSHA256 reports whether s is a structurally valid SCRAM-SHA-256
// secret in the form "SCRAM-SHA-256$<iter>:<salt>$<StoredKey>:<ServerKey>",
// matching parse_scram_secret in PostgreSQL.
func isSCRAMSHA256(s string) bool {
	parts := strings.Split(s, "$")
	if len(parts) != 3 {
		return false
	}
	if parts[0] != "SCRAM-SHA-256" {
		return false
	}

	iterSalt := strings.SplitN(parts[1], ":", 2)
	if len(iterSalt) != 2 {
		return false
	}
	iters, err := strconv.Atoi(iterSalt[0])
	if err != nil || iters < 1 {
		return false
	}
	if _, err := base64.StdEncoding.DecodeString(iterSalt[1]); err != nil {
		return false
	}

	keys := strings.SplitN(parts[2], ":", 2)
	if len(keys) != 2 {
		return false
	}
	storedKey, err := base64.StdEncoding.DecodeString(keys[0])
	if err != nil || len(storedKey) != scramSHA256KeyLen {
		return false
	}
	serverKey, err := base64.StdEncoding.DecodeString(keys[1])
	if err != nil || len(serverKey) != scramSHA256KeyLen {
		return false
	}
	return true
}

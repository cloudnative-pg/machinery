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
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/xdg-go/scram"

	"github.com/cloudnative-pg/machinery/pkg/postgres/password"
)

var (
	// ErrWrongComponents is returned when the hash is not split into the
	// three '$'-separated sections of the canonical SCRAM-SHA-256 form.
	ErrWrongComponents = errors.New("wrong number of components in password hash: expected 3 sections divided by '$'")

	// ErrWrongHashType is returned when the leading section of the hash is
	// not the literal "SCRAM-SHA-256".
	ErrWrongHashType = errors.New("wrong hash type (expected SCRAM-SHA-256)")

	// ErrWrongHashConfig is returned when the iter/salt section is not in
	// the expected "<iterations>:<salt>" form.
	ErrWrongHashConfig = errors.New(
		"wrong hash config (expected '<iterations>:<salt>' in the first '$' section)")

	// ErrWrongKeyComponents is returned when the key section is not in the
	// expected "<StoredKey>:<ServerKey>" form.
	ErrWrongKeyComponents = errors.New(
		"wrong key components (expected '<StoredKey>:<ServerKey>' in the last '$' section)")

	// ErrInvalidIterations is returned when the iteration count is not a
	// positive integer.
	ErrInvalidIterations = errors.New("iteration count must be a positive integer")

	// ErrInvalidStoredKey is returned when the StoredKey does not decode to
	// the SHA-256 digest size.
	ErrInvalidStoredKey = errors.New("stored key must decode to 32 bytes")

	// ErrInvalidServerKey is returned when the ServerKey does not decode to
	// the SHA-256 digest size.
	ErrInvalidServerKey = errors.New("server key must decode to 32 bytes")
)

// parsedHash contains the parsed PostgreSQL hash
type parsedHash struct {
	Iterations   int
	RawSalt      []byte
	RawStoredKey []byte
	RawServerKey []byte
}

// Verify checks if the passed SCRAM hash, in the format used by PostgreSQL,
// corresponds to the given plain text. It returns true on a match, false on
// mismatch, and a non-nil error only when hash is malformed.
//
// Verify performs PBKDF2 work proportional to the iteration count parsed
// from the hash, which the parser caps at 2^31-1 to match libpq.
// Callers that may receive attacker-influenced hashes should validate or
// cap the count further; PostgreSQL itself stores 4096 by default.
func Verify(hash string, plainText string) (bool, error) {
	parsedHash, err := parsePostgreSQLHash(hash)
	if err != nil {
		return false, fmt.Errorf("while parsing SCRAM hash: %w", err)
	}

	client, err := scram.SHA256.NewClient("", plainText, "")
	if err != nil {
		return false, fmt.Errorf("generating scram/SHA256 client: %w", err)
	}

	kf := scram.KeyFactors{
		Salt:  string(parsedHash.RawSalt),
		Iters: parsedHash.Iterations,
	}
	credentials, err := client.GetStoredCredentialsWithError(kf)
	if err != nil {
		return false, fmt.Errorf("computing stored credentials: %w", err)
	}

	computed := formatHash(parsedHash.RawSalt, parsedHash.Iterations, credentials.StoredKey, credentials.ServerKey)
	return subtle.ConstantTimeCompare([]byte(hash), []byte(computed)) == 1, nil
}

// parsePostgreSQLHash splits a PostgreSQL SCRAM secret of the form
// "SCRAM-SHA-256$<iter>:<salt>$<StoredKey>:<ServerKey>" into its components,
// returning one of the Err* sentinels for any structural error.
func parsePostgreSQLHash(hash string) (*parsedHash, error) {
	components := strings.Split(hash, "$")
	if len(components) != 3 {
		return nil, ErrWrongComponents
	}

	if components[0] != "SCRAM-SHA-256" {
		return nil, ErrWrongHashType
	}

	hashConfig := strings.Split(components[1], ":")
	if len(hashConfig) != 2 {
		return nil, ErrWrongHashConfig
	}

	keys := strings.Split(components[2], ":")
	if len(keys) != 2 {
		return nil, ErrWrongKeyComponents
	}

	// Match libpq's parse_scram_secret, which strtol's into a C int and
	// rejects ERANGE; cap at 32 bits so any hash a Postgres server would
	// accept is also accepted here, and no larger.
	iter64, err := strconv.ParseInt(hashConfig[0], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("while reading the number of iterations: %w", err)
	}
	iterations := int(iter64)
	if iterations < 1 {
		return nil, ErrInvalidIterations
	}

	rawSalt, err := base64.StdEncoding.DecodeString(hashConfig[1])
	if err != nil {
		return nil, fmt.Errorf("while base64-decoding salt: %w", err)
	}

	rawStoredKey, err := base64.StdEncoding.DecodeString(keys[0])
	if err != nil {
		return nil, fmt.Errorf("while base64-decoding stored key: %w", err)
	}
	if len(rawStoredKey) != password.SCRAMSHA256KeyLen {
		return nil, ErrInvalidStoredKey
	}

	rawServerKey, err := base64.StdEncoding.DecodeString(keys[1])
	if err != nil {
		return nil, fmt.Errorf("while base64-decoding server key: %w", err)
	}
	if len(rawServerKey) != password.SCRAMSHA256KeyLen {
		return nil, ErrInvalidServerKey
	}

	return &parsedHash{
		Iterations:   iterations,
		RawSalt:      rawSalt,
		RawStoredKey: rawStoredKey,
		RawServerKey: rawServerKey,
	}, nil
}

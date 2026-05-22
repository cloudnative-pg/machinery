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
	"fmt"

	"github.com/xdg-go/scram"
)

// DefaultPostgresIterations is the default number of PBKDF2 iterations used
// by PostgreSQL when hashing a SCRAM-SHA-256 secret, mirroring
// SCRAM_DEFAULT_ITERATIONS from PostgreSQL's src/include/common/scram-common.h.
const DefaultPostgresIterations = 4096

// DefaultSaltLength is the default raw-salt length used by PostgreSQL,
// mirroring SCRAM_DEFAULT_SALT_LEN from PostgreSQL's
// src/include/common/scram-common.h.
const DefaultSaltLength = 16

// GenerateOptions is the set of inputs to Generate.
type GenerateOptions struct {
	// Salt is the raw salt to be used. If empty, Generate uses a fresh
	// salt of DefaultSaltLength bytes drawn from crypto/rand.
	Salt []byte

	// Iterations is the PBKDF2 iteration count. If zero, Generate uses
	// DefaultPostgresIterations. A negative value is rejected with
	// ErrInvalidIterations.
	Iterations int

	// PlainText is the password to be hashed.
	PlainText string
}

// Generate returns a SCRAM hash for these options. It does not mutate
// the receiver, so repeated calls with Salt unset each draw a fresh
// salt.
func (options *GenerateOptions) Generate() (string, error) {
	local := *options
	if err := local.applyDefaults(); err != nil {
		return "", err
	}

	client, err := scram.SHA256.NewClient("", local.PlainText, "")
	if err != nil {
		return "", fmt.Errorf("generating scram/SHA256 client: %w", err)
	}

	kf := scram.KeyFactors{
		Salt:  string(local.Salt),
		Iters: local.Iterations,
	}
	credentials, err := client.GetStoredCredentialsWithError(kf)
	if err != nil {
		return "", fmt.Errorf("computing stored credentials: %w", err)
	}

	return formatHash(local.Salt, local.Iterations, credentials.StoredKey, credentials.ServerKey), nil
}

// applyDefaults mutates the receiver to fill in any unset option. It is
// unexported so callers cannot trigger the mutation themselves; Generate
// invokes it on a local copy.
func (options *GenerateOptions) applyDefaults() error {
	if options.Iterations < 0 {
		return ErrInvalidIterations
	}
	if options.Iterations == 0 {
		options.Iterations = DefaultPostgresIterations
	}

	if len(options.Salt) == 0 {
		rawSalt, err := makeSalt()
		if err != nil {
			return fmt.Errorf("while generating raw salt: %w", err)
		}
		options.Salt = rawSalt
	}

	return nil
}

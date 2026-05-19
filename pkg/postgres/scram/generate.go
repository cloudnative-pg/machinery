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

// DefaultPostgresIterations is the default number of iterations used by
// PostgreSQL
const DefaultPostgresIterations = 4096

// DefaultSaltLength is the default salt length as used by PostgreSQL
const DefaultSaltLength = 16

// GenerateOptions are information needed to generate a SCRAM hash
type GenerateOptions struct {
	// The raw salt to be used. If nil, a new salt of DefaultSaltLength bytes
	// will be automatically generated
	Salt []byte

	// The number of iterations. PostgreSQL uses 4096
	Iterations int

	// The plain password
	PlainText string
}

// Defaults fills the default values into the options if they
// have not have been already defined
func (options *GenerateOptions) Defaults() error {
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

// Generate generates a SCRAM hash from the options. Missing fields are
// populated by Defaults.
func (options *GenerateOptions) Generate() (string, error) {
	if err := options.Defaults(); err != nil {
		return "", err
	}

	client, err := scram.SHA256.NewClient("", options.PlainText, "")
	if err != nil {
		return "", fmt.Errorf("generating scram/SHA256 client: %w", err)
	}

	kf := scram.KeyFactors{
		Salt:  string(options.Salt),
		Iters: options.Iterations,
	}
	credentials, err := client.GetStoredCredentialsWithError(kf)
	if err != nil {
		return "", fmt.Errorf("computing stored credentials: %w", err)
	}

	return formatHash(options.Salt, options.Iterations, credentials.StoredKey, credentials.ServerKey), nil
}

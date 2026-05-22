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

// Package scram generates and verifies SCRAM-SHA-256 password hashes in the
// form PostgreSQL stores them in pg_authid.rolpassword, namely
// "SCRAM-SHA-256$<iter>:<salt>$<StoredKey>:<ServerKey>".
//
// It is concerned with the on-disk representation of the secret only; it
// does not implement the SCRAM SASL authentication exchange between client
// and server.
package scram

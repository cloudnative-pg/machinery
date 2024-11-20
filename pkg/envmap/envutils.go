/*
Copyright The CloudNativePG Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package envmap

import (
	"fmt"
	"maps"
	"os"
	"strings"

	"github.com/cloudnative-pg/machinery/pkg/stringset"
)

// ErrWrongEnvironmentString is raised when the Parse function detects
// a wrong environment string.
// By convention, the strings in environment should have the form
// "name=value".
//
// See: https://www.man7.org/linux/man-pages/man7/environ.7.html
type ErrWrongEnvironmentString struct {
	entry string
}

// Error implements the error interface.
func (e *ErrWrongEnvironmentString) Error() string {
	return fmt.Sprintf("wrong environment variable entry: %s", e.entry)
}

// EnvironmentMap represent a map between environment variable names
// and their value.
type EnvironmentMap map[string]string

// Parse parses a list of strings in the form "foo=bar" into an environment map.
// If an envalid string is detected, this function will return a
// ErrWrongEnvironmentString error for the invalid entry.
func Parse(env []string) (EnvironmentMap, error) {
	result := make(map[string]string, len(env))

	for _, entry := range env {
		prefix, suffix, found := strings.Cut(entry, "=")
		if !found {
			return nil, &ErrWrongEnvironmentString{
				entry: entry,
			}
		}

		result[prefix] = suffix
	}

	return result, nil
}

// StringSlice converts an environment map to a list of strings in the form "foo=bar".
// The returned list is sorted in lexicographic key order.
func (e EnvironmentMap) StringSlice() []string {
	keys := stringset.FromKeys(e).ToSortedList()
	result := make([]string, len(keys))
	for i, keyName := range keys {
		result[i] = fmt.Sprintf("%s=%s", keyName, e[keyName])
	}
	return result
}

// ParseEnviron returns the environment of the current process.
func ParseEnviron() (EnvironmentMap, error) {
	return Parse(os.Environ())
}

// Merge merges two environment maps.
// When a key in e1 is also present in e2,
// the value associated with the key in e2 will be used.
func Merge(e1, e2 EnvironmentMap) EnvironmentMap {
	result := make(EnvironmentMap, len(e1)+len(e2))
	maps.Copy(result, e1)
	maps.Copy(result, e2)
	return result
}

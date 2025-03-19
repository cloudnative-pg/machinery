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

package pgconfig

import (
	"fmt"
	"os/exec"
	"strings"
)

// InstallationLocation is the type of the PostgreSQL
// installation locations
type InstallationLocation string

const (
	// BinDir is the location user executables. Use this, for example,
	// to find the psql program. This is normally also the location
	// where the pg_config program resides.
	BinDir InstallationLocation = "bindir"

	// PkgLibDir is the location of dynamically loadable modules, or
	// where the server would search for them. (Other
	// architecture-dependent data files might also be installed in
	// this directory.)
	PkgLibDir InstallationLocation = "pkglibdir"

	// ShareDir is the location of architecture-independent support
	// files.
	ShareDir InstallationLocation = "sharedir"
)

// GetPgConfigDirectory retrieves a PostgreSQL directory path using the
// specified InstallationLocation
func GetPgConfigDirectory(pgConfigBinary string, loc InstallationLocation) (string, error) {
	out, err := exec.Command(pgConfigBinary, "--"+string(loc)).Output() //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("failed to get the %q value from pg_config: %w", loc, err)
	}
	return strings.TrimSpace(string(out)), nil
}

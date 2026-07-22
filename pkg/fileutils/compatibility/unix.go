//go:build linux || darwin

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

// Package compatibility provides a layer to cross-compile with other OS than Linux
package compatibility

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

// CreateFifo invokes the Unix system call Mkfifo, if the given filename
// doesn't already exist. If a filesystem entry already exists at that path,
// it must already be a FIFO: any other type is reported as an error rather
// than silently left in place.
func CreateFifo(fileName string) error {
	isFifo := func(fileMode os.FileMode) bool {
		return fileMode&os.ModeNamedPipe != 0
	}

	info, err := os.Lstat(fileName)
	switch {
	case err == nil:
		if !isFifo(info.Mode()) {
			return fmt.Errorf("%q: %w", fileName, ErrExistsNotFifo)
		}
		return nil
	case os.IsNotExist(err):
		return unix.Mkfifo(fileName, 0o600)
	default:
		return err
	}
}

// AddInstanceRunCommands adds specific OS commands to the postgres exec.Cmd
func AddInstanceRunCommands(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// Umask sets the process's unix umask to prevent/allow permissions changes
func Umask(mask int) int {
	return unix.Umask(mask)
}

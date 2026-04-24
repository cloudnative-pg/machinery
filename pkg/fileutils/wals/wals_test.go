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

package wals

import (
	"context"
	"os"
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("WALList functions", func() {
	var walList *WALList
	var ctx context.Context
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "wal_test")
		Expect(err).ToNot(HaveOccurred())

		walList = &WALList{
			pgDataPath: tmpDir,
			Ready:      []string{"000000010000000000000001", "000000010000000000000002"},
			Done:       []string{},
		}
		ctx = context.TODO()

		// Create the .ready files
		archiveStatusPath := filepath.Join(tmpDir, "pg_wal", "archive_status")
		err = os.MkdirAll(archiveStatusPath, 0o750)
		Expect(err).ToNot(HaveOccurred())

		for _, walName := range walList.Ready {
			readyFilePath := filepath.Join(archiveStatusPath, walName+".ready")
			file, err := os.Create(readyFilePath) // nolint:gosec
			Expect(err).ToNot(HaveOccurred())
			err = file.Close()
			Expect(err).ToNot(HaveOccurred())
		}
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpDir)
		Expect(err).ToNot(HaveOccurred())
	})

	It("removes a ready item", func() {
		walList.RemoveReadyItem("000000010000000000000001")
		Expect(walList.Ready).To(Equal([]string{"000000010000000000000002"}))
	})

	It("returns ready items as a slice", func() {
		readyItems := walList.ReadyItemsToSlice()
		Expect(readyItems).To(Equal([]string{"000000010000000000000001", "000000010000000000000002"}))
	})

	It("marks a WAL file as done", func() {
		err := walList.MarkAsDone(ctx, "000000010000000000000001")
		Expect(err).ToNot(HaveOccurred())
		Expect(walList.Ready).To(Equal([]string{"000000010000000000000002"}))
		Expect(walList.Done).To(Equal([]string{"000000010000000000000001"}))
	})

	It("gathers ready WAL files", func() {
		result := GatherReadyWALFiles(ctx, GatherReadyWALFilesConfig{MaxResults: 10, PgDataPath: tmpDir})
		Expect(result.Ready).To(
			ContainElement(
				path.Join(tmpDir, "pg_wal/000000010000000000000001")))
		Expect(
			result.Ready).To(
			ContainElement(path.Join(tmpDir, "pg_wal/000000010000000000000002")))
		Expect(result.HasMoreResults).To(BeFalse())
	})

	It("gathers ready WAL files with one skipped", func() {
		result := GatherReadyWALFiles(ctx, GatherReadyWALFilesConfig{
			MaxResults: 10, PgDataPath: tmpDir, SkipWALs: []string{"pg_wal/000000010000000000000001"},
		})
		Expect(result.Ready).ToNot(
			ContainElement(
				path.Join(tmpDir, "pg_wal/000000010000000000000001")))
		Expect(
			result.Ready).To(
			ContainElement(path.Join(tmpDir, "pg_wal/000000010000000000000002")))
		Expect(result.HasMoreResults).To(BeFalse())
	})

	It("handles no more WAL files needed", func() {
		result := GatherReadyWALFiles(ctx, GatherReadyWALFilesConfig{MaxResults: 1, PgDataPath: tmpDir})
		Expect(result.Ready).To(HaveLen(1))
		Expect(result.HasMoreResults).To(BeTrue())
	})
})

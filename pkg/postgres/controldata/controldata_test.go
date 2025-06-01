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

package controldata

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const fakeControlData = `pg_control version number:               1002
Catalog version number:                  202201241
Database cluster state:                  shut down
Database system identifier:              12345678901234567890123456789012
Latest checkpoint's TimeLineID:       3
pg_control last modified:                2024-04-30 12:00:00 UTC
Latest checkpoint location:              0/3000FF0
Prior checkpoint location:               0/2000AA0
Minimum recovery ending location:        0/3000000
Time of latest checkpoint:               2024-04-30 10:00:00 UTC
Database block size:                     8192 bytes
Latest checkpoint's REDO location:         0/3000CC0
Latest checkpoint's REDO WAL file:         000000010000000000000003
Blocks per segment of large relation:    131072
Maximum data alignment:                  8
Database disk usage:                     10240 KB
Maximum xlog ID:                         123456789
Next xlog byte position:                 0/3000010`

const fakeWrongControlData = `pg_control version number:               1002
Catalog version number:                  202201241
Database cluster state:                  shut down
Database system identifier:              12345678901234567890123456789012
Latest checkpoint's TimeLineID:       3
pg_control last modified:                2024-04-30 12:00:00 UTC
Latest checkpoint location:              0/3000FF0
Prior checkpoint location:               0/2000AA0
THIS IS A TEST!
Minimum recovery ending location:        0/3000000
Time of latest checkpoint:               2024-04-30 10:00:00 UTC
Database block size:                     8192 bytes
Latest checkpoint's REDO location:         0/3000CC0
Latest checkpoint's REDO WAL file:         000000010000000000000003
Blocks per segment of large relation:    131072
Maximum data alignment:                  8
Database disk usage:                     10240 KB
Maximum xlog ID:                         123456789
Next xlog byte position:                 0/3000010`

var _ = DescribeTable("PGData database state parser",
	func(ctx SpecContext, state string, isShutDown bool) {
		Expect(PgDataState(state).IsShutdown(ctx)).To(Equal(isShutDown))
	},
	Entry("A primary PostgreSQL instance has been shut down", "shut down", true),
	Entry("A standby PostgreSQL instance has been shut down", "shut down in recovery", true),
	Entry("A primary instance is up and running", "in production", false),
	Entry("A standby instance is up and running", "in archive recovery", false),
	Entry("An unknown state", "unknown-state", false),
)

var _ = Describe("pg_controldata output parser", func() {
	It("parse a correct output", func() {
		fakeControlDataEntries := len(strings.Split(fakeControlData, "\n"))
		output := ParseOutput(fakeControlData)
		Expect(output["Catalog version number"]).To(Equal("202201241"))
		Expect(output["Database disk usage"]).To(Equal("10240 KB"))
		Expect(output).To(HaveLen(fakeControlDataEntries))
	})

	It("silently skips wrong lines", func() {
		correctOutput := ParseOutput(fakeControlData)
		wrongOutput := ParseOutput(fakeWrongControlData)
		Expect(correctOutput).To(Equal(wrongOutput))
	})

	It("returns an empty map when the output is empty", func() {
		output := ParseOutput("")
		Expect(output).To(BeEmpty())
	})
})

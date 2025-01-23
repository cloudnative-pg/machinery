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

package types

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LSN handling functions", func() {
	Describe("Parse and Int64ToLSN", func() {
		It("raises errors for invalid LSNs", func() {
			_, err := LSN("").Parse()
			Expect(err).To(HaveOccurred())
			_, err = LSN("/").Parse()
			Expect(err).To(HaveOccurred())
			_, err = LSN("28734982739847293874823974928738423/987429837498273498723984723").Parse()
			Expect(err).To(HaveOccurred())
		})

		DescribeTable(
			"works for good LSNs",
			func(lsn string, value uint64) {
				Expect(LSN(lsn).Parse()).To(Equal(value))
				Expect(Int64ToLSN(value)).To(Equal(LSN(lsn)))
			},
			Entry("1/1", "1/1", uint64(4294967297)),
			Entry("3/23", "3/23", uint64(12884901923)),
			Entry("3BB/A9FFFBE8", "3BB/A9FFFBE8", uint64(4104545893352)),
		)
	})

	Describe("Less", func() {
		It("handles errors in the same way as the zero LSN value", func() {
			Expect(LSN("").Less("3/23")).To(BeTrue())
			Expect(LSN("3/23").Less("")).To(BeFalse())
		})

		It("works correctly for good LSNs", func() {
			Expect(LSN("1/23").Less(LSN("1/24"))).To(BeTrue())
			Expect(LSN("1/24").Less(LSN("1/23"))).To(BeFalse())
			Expect(LSN("1/23").Less(LSN("2/23"))).To(BeTrue())
			Expect(LSN("2/23").Less(LSN("1/23"))).To(BeFalse())
		})
	})

	Describe("XLog file name", func() {
		segmentSize := uint64(16 * 1024 * 1024)

		It("raise errors for invalid LSNs", func() {
			_, err := LSN("").WALFileName(1, segmentSize)
			Expect(err).To(HaveOccurred())
		})

		DescribeTable(
			"works correctly for good LSNs",
			func(tli int, lsn LSN, wal string) {
				Expect(lsn.WALFileName(tli, segmentSize)).To(Equal(wal))

				lsnWalStart, err := lsn.WALFileStart(segmentSize)
				Expect(err).ToNot(HaveOccurred())
				Expect(lsnWalStart.WALFileName(tli, segmentSize)).To(Equal(wal))
			},
			Entry("good LSN", 10, LSN("5283/D9C2A320"), "0000000A00005283000000D9"),
			Entry("good LSN", 1, LSN("C/CE7BAD70"), "000000010000000C000000CE"),
			Entry("good LSN", 1, LSN("0/14E5EB8"), "000000010000000000000001"),
		)
	})
})

var _ = Describe("LSNStartFromWALName", func() {
	segmentSize := uint64(16 * 1024 * 1024)

	It("returns an error for invalid WAL file name length", func() {
		_, err := LSNStartFromWALName("00000001000000000000001", segmentSize)
		Expect(err).To(HaveOccurred())
	})

	It("returns an error for invalid WAL file name format", func() {
		_, err := LSNStartFromWALName("invalidWALFileName", segmentSize)
		Expect(err).To(HaveOccurred())
	})

	It("returns the correct start LSN for a valid WAL file name", func() {
		startLSN, err := LSNStartFromWALName("000000010000000000000001", segmentSize)
		Expect(err).ToNot(HaveOccurred())
		Expect(startLSN).To(Equal(LSN("0/1000000")))
	})

	It("returns the correct start LSN for another valid WAL file name", func() {
		startLSN, err := LSNStartFromWALName("0000000A00005283000000D9", segmentSize)
		Expect(err).ToNot(HaveOccurred())
		Expect(startLSN).To(Equal(LSN("5283/D9000000")))
	})

	// Additional test cases
	It("returns the correct start LSN for a different valid WAL file name", func() {
		startLSN, err := LSNStartFromWALName("000000020000003400000056", segmentSize)
		Expect(err).ToNot(HaveOccurred())

		const lsn = LSN("34/56000000")
		Expect(startLSN).To(Equal(lsn))
		Expect(lsn.WALFileStart(segmentSize)).To(Equal(lsn))
	})

	It("returns the correct start LSN for a WAL file name with different timeline", func() {
		startLSN, err := LSNStartFromWALName("0000000B0000001C0000002A", segmentSize)
		Expect(err).ToNot(HaveOccurred())

		const lsn = LSN("1C/2A000000")
		Expect(startLSN).To(Equal(lsn))
		Expect(lsn.WALFileStart(segmentSize)).To(Equal(lsn))
	})

	It("returns the correct start LSN for a WAL file name with maximum segment", func() {
		const filename = "FFFFFFFFFFFFFFFFFFFFFFFF"
		const lsn = LSN("FFFFFE/FF000000")

		startLSN, err := LSNStartFromWALName(filename, segmentSize)
		Expect(err).ToNot(HaveOccurred())
		Expect(startLSN).To(Equal(lsn))

		Expect(lsn.WALFileStart(segmentSize)).To(Equal(lsn))
	})
})

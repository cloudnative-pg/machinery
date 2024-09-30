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

package version

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PostgreSQL version handling", func() {
	Describe("parsing", func() {
		It("should parse versions >= 10", func() {
			Expect(FromTag("10.3")).To(Equal(New(10, 3)))
			Expect(FromTag("12.3")).To(Equal(New(12, 3)))
		})

		It("should parse PostgreSQL official version policy", func() {
			Expect(FromTag("11.2")).To(Equal(New(11, 2)))
			Expect(FromTag("12.1")).To(Equal(New(12, 1)))
			Expect(FromTag("13.3.2.1-1")).To(Equal(New(13, 3)))
			Expect(FromTag("13.4")).To(Equal(New(13, 4)))
			Expect(FromTag("14")).To(Equal(New(14, 0)))
			Expect(FromTag("15.5-10")).To(Equal(New(15, 5)))
			Expect(FromTag("16.0")).To(Equal(New(16, 0)))
			Expect(FromTag("17beta1")).To(Equal(New(17, 0)))
			Expect(FromTag("17rc1")).To(Equal(New(17, 0)))
		})

		It("should ignore extra components", func() {
			Expect(FromTag("3.4.3.2.5")).To(Equal(New(3, 4)))
			Expect(FromTag("10.11.12")).To(Equal(New(10, 11)))
			Expect(FromTag("9.4_beautiful")).To(Equal(New(9, 4)))
			Expect(FromTag("11-1")).To(Equal(New(11, 0)))
			Expect(FromTag("15beta1")).To(Equal(New(15, 0)))
		})

		It("should gracefully handle errors", func() {
			_, err := FromTag("")
			Expect(err).To(HaveOccurred())

			_, err = FromTag("10.five")
			Expect(err).To(HaveOccurred())

			_, err = FromTag("11.old")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("detect whenever a version upgrade is possible using the numeric version", func() {
		It("succeed when the major version is the same", func() {
			Expect(IsUpgradePossible(New(10, 0), New(10, 3))).To(BeTrue())
		})

		It("prevent upgrading to a different major version", func() {
			Expect(IsUpgradePossible(New(10, 3), New(11, 3))).To(BeFalse())
		})
	})

	Describe("comparison operator between versions", func() {
		It("compares two versions for equality", func() {
			Expect(New(12, 2)).To(Equal(New(12, 2)))

			v := New(10, 3)
			Expect(v).To(Equal(v))
		})

		DescribeTable(
			"'less than' operator",
			func(a, b Data, result bool) {
				Expect(a.Less(b)).To(Equal(result))
				Expect(a.Less(a)).To(BeFalse())
				Expect(b.Less(b)).To(BeFalse())

				if a == b {
					Expect(a.Less(b)).To(Equal(false))
					Expect(b.Less(a)).To(Equal(false))
				} else {
					Expect(b.Less(a)).To(Equal(!result))
				}

				if !a.Less(b) && !b.Less(a) {
					Expect(a).To(Equal(b))
				}
			},
			Entry("same major: 12.3 vs 12.2", New(12, 3), New(12, 2), false),
			Entry("same major: 12.2 vs 12.3", New(12, 2), New(12, 3), true),
			Entry("different major: 12.3 vs 13.4", New(12, 3), New(13, 4), true),
			Entry("different major: 13.4 vs 12.3", New(13, 4), New(12, 3), false),
			Entry("different major: 12.4 vs 13.3", New(12, 4), New(13, 3), true),
			Entry("different major: 13.3 vs 12.4", New(13, 3), New(12, 4), false),
			Entry("equal: 12.2 vs 12.2", New(12, 2), New(12, 2), false),
		)
	})
})

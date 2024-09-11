package types

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parsing targetTime", func() {
	It("parsing works given targetTime in `YYYY-MM-DD HH24:MI:SS` format", func() {
		res, err := ParseTargetTime(nil, "2021-09-01 10:22:47")
		Expect(err).ToNot(HaveOccurred())
		Expect(res.MarshalText()).To(BeEquivalentTo("2021-09-01T10:22:47Z"))
	})

	It("parsing works given targetTime in `YYYY-MM-DD HH24:MI:SS.FF6TZH` format", func() {
		res, err := ParseTargetTime(nil, "2021-09-01 10:22:47.000000+06")
		Expect(err).ToNot(HaveOccurred())
		Expect(res.MarshalText()).To(BeEquivalentTo("2021-09-01T10:22:47+06:00"))
	})

	It("parsing works given targetTime in `YYYY-MM-DD HH24:MI:SS.FF6TZH:TZM` format", func() {
		res, err := ParseTargetTime(nil, "2021-09-01 10:22:47.000000+06:00")
		Expect(err).ToNot(HaveOccurred())
		Expect(res.MarshalText()).To(BeEquivalentTo("2021-09-01T10:22:47+06:00"))
	})

	It("parsing works given targetTime in `YYYY-MM-DDTHH24:MI:SSZ` format", func() {
		res, err := ParseTargetTime(nil, "2021-09-01T10:22:47Z")
		Expect(err).ToNot(HaveOccurred())
		Expect(res.MarshalText()).To(BeEquivalentTo("2021-09-01T10:22:47Z"))
	})

	It("parsing works given targetTime in `YYYY-MM-DDTHH24:MI:SS±TZH:TZM` format", func() {
		res, err := ParseTargetTime(nil, "2021-09-01T10:22:47+00:00")
		Expect(err).ToNot(HaveOccurred())
		Expect(res.MarshalText()).To(BeEquivalentTo("2021-09-01T10:22:47Z"))
	})

	It("parsing works given targetTime in `YYYY-MM-DDTHH24:MI:SS` format", func() {
		res, err := ParseTargetTime(nil, "2021-09-01T10:22:47")
		Expect(err).ToNot(HaveOccurred())
		Expect(res.MarshalText()).To(BeEquivalentTo("2021-09-01T10:22:47Z"))
	})

	It("parsing works with RFC3339Micro format `YYYY-MM-DDTHH24:MI:SS.SSSSSSZ`", func() {
		_, err := ParseTargetTime(nil, "2006-01-02T15:04:05.000000Z")
		Expect(err).ToNot(HaveOccurred())
	})
})

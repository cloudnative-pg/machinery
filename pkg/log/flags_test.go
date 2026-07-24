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

package log

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("customDestination", func() {
	AfterEach(func() {
		logDestination = ""
		truncateDestination = false
	})

	It("does nothing when no log destination is configured", func() {
		options := &zap.Options{}
		customDestination(options)
		Expect(options.DestWriter).To(BeNil())
	})

	It("appends to an existing regular file instead of overwriting from offset 0", func() {
		destPath := filepath.Join(GinkgoT().TempDir(), "log.txt")
		Expect(os.WriteFile(destPath, []byte("0123456789"), 0o600)).To(Succeed())

		logDestination = destPath
		options := &zap.Options{}
		customDestination(options)
		Expect(options.DestWriter).NotTo(BeNil())

		_, err := options.DestWriter.Write([]byte("abc"))
		Expect(err).NotTo(HaveOccurred())

		if closer, ok := options.DestWriter.(io.Closer); ok {
			Expect(closer.Close()).To(Succeed())
		}

		content, err := os.ReadFile(destPath) //#nosec
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("0123456789abc"))
	})

	It("keeps appending across repeated opens, as would happen across separate process invocations", func() {
		destPath := filepath.Join(GinkgoT().TempDir(), "log.txt")
		logDestination = destPath

		firstOpen := &zap.Options{}
		customDestination(firstOpen)
		_, err := firstOpen.DestWriter.Write([]byte("first-run\n"))
		Expect(err).NotTo(HaveOccurred())
		Expect(firstOpen.DestWriter.(io.Closer).Close()).To(Succeed())

		secondOpen := &zap.Options{}
		customDestination(secondOpen)
		_, err = secondOpen.DestWriter.Write([]byte("second\n"))
		Expect(err).NotTo(HaveOccurred())
		Expect(secondOpen.DestWriter.(io.Closer).Close()).To(Succeed())

		content, err := os.ReadFile(destPath) //#nosec
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("first-run\nsecond\n"))
	})

	It("truncates the destination when SetTruncateDestination(true) was called", func() {
		destPath := filepath.Join(GinkgoT().TempDir(), "log.txt")
		Expect(os.WriteFile(destPath, []byte("0123456789"), 0o600)).To(Succeed())

		SetTruncateDestination(true)
		defer SetTruncateDestination(false)

		logDestination = destPath
		options := &zap.Options{}
		customDestination(options)
		Expect(options.DestWriter).NotTo(BeNil())

		_, err := options.DestWriter.Write([]byte("abc"))
		Expect(err).NotTo(HaveOccurred())
		Expect(options.DestWriter.(io.Closer).Close()).To(Succeed())

		content, err := os.ReadFile(destPath) //#nosec
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("abc"))
	})
})

// configureTestLogging runs the production configuration path: it binds the
// logging flags, parses them, and calls ConfigureLogging, sending the log
// stream to dest so the specs can inspect it.
func configureTestLogging(dest string, extraFlags []string, opts ...ConfigureOption) {
	flags := &Flags{}
	flagSet := &pflag.FlagSet{}
	flags.AddFlags(flagSet)
	args := append([]string{"--log-destination", dest}, extraFlags...)
	ExpectWithOffset(1, flagSet.Parse(args)).To(Succeed())
	flags.ConfigureLogging(opts...)
}

// destLines returns the lines of the log destination file containing marker
func destLines(dest string, marker string) []string {
	content, err := os.ReadFile(dest) //nolint:gosec
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	var result []string
	for _, line := range strings.Split(string(content), "\n") {
		if strings.Contains(line, marker) {
			result = append(result, line)
		}
	}
	return result
}

var _ = Describe("ConfigureLogging sampling behavior", func() {
	// burst must exceed the 100 msgs/s initial-pass threshold of the sampler
	// installed by the controller-runtime zap builder
	const burst = 300

	var dest string

	BeforeEach(func() {
		dest = filepath.Join(GinkgoT().TempDir(), "log")
	})

	AfterEach(func() {
		logDestination = ""
	})

	It("keeps the duplicate-message sampler by default", func() {
		configureTestLogging(dest, nil)

		for range burst {
			Info("sampled-burst-marker")
		}

		lines := destLines(dest, `"msg":"sampled-burst-marker"`)
		// A single sampler window passes 102 of the 300 identical messages
		// (the first 100, then 1 in 100). If the loop straddles a window
		// boundary, the worst-case split passes 201. Only a pathological
		// scheduler stall spreading the burst over three or more windows
		// could exceed this bound.
		Expect(len(lines)).To(BeNumerically("<=", 210),
			"the default logger should sample duplicate messages beyond 100/s")
	})

	It("emits every record during a burst when sampling is disabled", func() {
		configureTestLogging(dest, nil, WithDisabledSampling())

		for range burst {
			Info("unsampled-burst-marker")
		}

		lines := destLines(dest, `"msg":"unsampled-burst-marker"`)
		Expect(lines).To(HaveLen(burst))
	})

	It("still honors --log-level when sampling is disabled", func() {
		configureTestLogging(dest, []string{"--log-level", "error"}, WithDisabledSampling())

		Info("filtered-info-marker")
		Error(errors.New("boom"), "error-marker")

		Expect(destLines(dest, "filtered-info-marker")).To(BeEmpty(),
			"info records must still be filtered out at the error level")
		Expect(destLines(dest, "error-marker")).To(HaveLen(1))
	})

	It("keeps the trace level fully functional when sampling is disabled", func() {
		// at debug/trace the controller-runtime builder never installs the
		// sampler, so here the level restoration wraps a core that is
		// already at the requested level
		configureTestLogging(dest, []string{"--log-level", "trace"}, WithDisabledSampling())

		Trace("trace-marker")
		Info("info-marker")

		Expect(destLines(dest, "trace-marker")).To(HaveLen(1))
		Expect(destLines(dest, "info-marker")).To(HaveLen(1))
	})

	It("still honors the field remapping flags when sampling is disabled", func() {
		configureTestLogging(dest,
			[]string{"--log-field-level", "severity", "--log-field-timestamp", "event_time"},
			WithDisabledSampling())

		Info("remap-marker")

		lines := destLines(dest, "remap-marker")
		Expect(lines).To(HaveLen(1))
		Expect(lines[0]).To(ContainSubstring(`"severity":"info"`))
		Expect(lines[0]).To(ContainSubstring(`"event_time":`))
	})
})

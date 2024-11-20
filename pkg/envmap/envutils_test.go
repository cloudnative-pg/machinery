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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Environment map", func() {
	DescribeTable(
		"Environment map parsing",
		func(env []string, expectedResult EnvironmentMap, expectedError error) {
			result, err := Parse(env)
			if expectedError == nil {
				Expect(err).ToNot(HaveOccurred())
			} else {
				Expect(err).To(Equal(expectedError))
			}
			Expect(result).To(Equal(expectedResult))
		},
		Entry(
			"nil slice",
			nil,
			map[string]string{},
			nil,
		),
		Entry(
			"basic test",
			[]string{
				"PATH=/usr/local/bin:/usr/bin",
				"TERM=xterm-256color",
			},
			map[string]string{
				"PATH": "/usr/local/bin:/usr/bin",
				"TERM": "xterm-256color",
			},
			nil,
		),
		Entry(
			"duplicate key",
			[]string{
				"PATH=/usr/local/bin:/usr/bin",
				"PATH=/usr/bin",
			},
			map[string]string{
				"PATH": "/usr/bin",
			},
			nil,
		),
		Entry(
			"wrong entry (too many equal signs)",
			[]string{
				"PATH=/usr/local/bin:/usr/bin",
				"TERM=xterm-256color=boh",
			},
			map[string]string{
				"PATH": "/usr/local/bin:/usr/bin",
				"TERM": "xterm-256color=boh",
			},
			nil,
		),
		Entry(
			"wrong entry (no equal sign)",
			[]string{
				"PATH=/usr/local/bin:/usr/bin",
				"TERM",
			},
			nil,
			&ErrWrongEnvironmentString{
				entry: "TERM",
			},
		),
	)

	It("parses the current environment", func() {
		envMap, err := ParseEnviron()
		Expect(err).ToNot(HaveOccurred())
		Expect(envMap).ToNot(BeNil())
	})

	DescribeTable(
		"Environment map merging",
		func(e1, e2, result EnvironmentMap) {
			Expect(Merge(e1, e2)).To(Equal(result))
		},
		Entry(
			"nil arguments",
			nil,
			nil,
			map[string]string{},
		),
		Entry(
			"e1 is nil",
			nil,
			map[string]string{
				"PATH": "/usr/local/bin",
			},
			map[string]string{
				"PATH": "/usr/local/bin",
			},
		),
		Entry(
			"e2 is nil",
			map[string]string{
				"PATH": "/usr/local/bin",
			},
			nil,
			map[string]string{
				"PATH": "/usr/local/bin",
			},
		),
		Entry(
			"e1 and e2 have no common keys",
			map[string]string{
				"PATH": "/usr/local/bin",
			},
			map[string]string{
				"TERM": "xterm-256color",
			},
			map[string]string{
				"PATH": "/usr/local/bin",
				"TERM": "xterm-256color",
			},
		),
		Entry(
			"e1 and e2 have common keys, e2 will override e1",
			map[string]string{
				"PATH": "/usr/local/bin",
				"TERM": "xterm-256color",
			},
			map[string]string{
				"PATH": "/etc",
			},
			map[string]string{
				"PATH": "/etc",
				"TERM": "xterm-256color",
			},
		),
	)

	DescribeTable(
		"Environment map StringSlice",
		func(e EnvironmentMap, expectedResult []string) {
			Expect(e.StringSlice()).To(Equal(expectedResult))
		},
		Entry(
			"nil environment map",
			nil,
			[]string{},
		),
		Entry(
			"empty environment map",
			map[string]string{},
			[]string{},
		),
		Entry(
			"environment map sorted in lexicographical order",
			map[string]string{
				"TERM":    "xterm-256color",
				"PATH":    "/usr/local/bin",
				"GPG_TTY": "/dev/ttys008",
			},
			[]string{
				"GPG_TTY=/dev/ttys008",
				"PATH=/usr/local/bin",
				"TERM=xterm-256color",
			},
		),
	)
})

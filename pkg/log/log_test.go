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
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-logr/logr"
	uberzap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// newJSONLogger returns a logger backed by zap that writes the entries
// enabled at the given level, as JSON, to the returned buffer
func newJSONLogger(level zapcore.Level) (*logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return &logger{Logger: zap.New(zap.WriteTo(&buf), zap.Level(level))}, &buf
}

// entries decodes the JSON entries written to the buffer
func entries(buf *bytes.Buffer) []map[string]interface{} {
	var result []map[string]interface{}
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var entry map[string]interface{}
		Expect(json.Unmarshal([]byte(line), &entry)).To(Succeed())
		result = append(result, entry)
	}
	return result
}

var _ = Describe("Warning", func() {
	It("logs at warning level", func() {
		l, buf := newJSONLogger(DebugLevel)

		l.Warning("a warning", "key", "value")

		Expect(entries(buf)).To(ConsistOf(SatisfyAll(
			HaveKeyWithValue("level", "warn"),
			HaveKeyWithValue("msg", "a warning"),
			HaveKeyWithValue("key", "value"),
		)))
	})

	It("keeps the values and the name of the logger", func() {
		l, buf := newJSONLogger(DebugLevel)

		l.WithName("component").WithValues("cluster", "example").Warning("a warning")

		Expect(entries(buf)).To(ConsistOf(SatisfyAll(
			HaveKeyWithValue("logger", "component"),
			HaveKeyWithValue("cluster", "example"),
		)))
	})

	It("reports the caller of Warning", func() {
		var buf bytes.Buffer
		l := &logger{Logger: zap.New(
			zap.WriteTo(&buf),
			zap.Level(DebugLevel),
			zap.RawZapOpts(uberzap.AddCaller()),
		)}

		_, file, line, _ := runtime.Caller(0)
		l.Warning("a warning")

		expected := fmt.Sprintf("%s/%s:%d", filepath.Base(filepath.Dir(file)), filepath.Base(file), line+1)
		Expect(entries(&buf)).To(ConsistOf(HaveKeyWithValue("caller", expected)))
	})

	It("is logged when the level is warning, unlike info", func() {
		l, buf := newJSONLogger(WarningLevel)

		l.Info("an info")
		l.Warning("a warning")

		Expect(entries(buf)).To(ConsistOf(SatisfyAll(
			HaveKeyWithValue("level", "warn"),
			HaveKeyWithValue("msg", "a warning"),
		)))
	})

	It("is not logged when the level is error", func() {
		l, buf := newJSONLogger(ErrorLevel)

		l.Warning("a warning")

		Expect(entries(buf)).To(BeEmpty())
	})

	It("is logged at warning level by the package-level function", func() {
		previous := log.Logger
		DeferCleanup(func() { SetLogger(previous) })

		l, buf := newJSONLogger(WarningLevel)
		SetLogger(l.Logger)

		Warning("a warning")

		Expect(entries(buf)).To(ConsistOf(HaveKeyWithValue("level", "warn")))
	})

	It("falls back to info for sinks not supporting slog", func() {
		sink := &infoSink{}
		l := &logger{Logger: logr.New(sink)}

		l.Warning("a warning")

		Expect(sink.messages).To(ConsistOf("a warning"))
		Expect(sink.levels).To(ConsistOf(0))
	})
})

// infoSink is a logr.LogSink not supporting slog, recording the messages
// logged at info level and their verbosity levels
type infoSink struct {
	messages []string
	levels   []int
}

func (s *infoSink) Init(logr.RuntimeInfo) {}

func (s *infoSink) Enabled(int) bool { return true }

func (s *infoSink) Info(level int, msg string, _ ...interface{}) {
	s.messages = append(s.messages, msg)
	s.levels = append(s.levels, level)
}

func (s *infoSink) Error(error, string, ...interface{}) {}

func (s *infoSink) WithValues(...interface{}) logr.LogSink { return s }

func (s *infoSink) WithName(string) logr.LogSink { return s }

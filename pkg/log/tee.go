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

import "github.com/go-logr/logr"

// teeSink sends log records to two logr.LogSink implementations.
type teeSink struct {
	primary, secondary logr.LogSink
}

func (t *teeSink) Init(info logr.RuntimeInfo) {
	if t.primary != nil {
		t.primary.Init(info)
	}
	if t.secondary != nil {
		t.secondary.Init(info)
	}
}

func (t *teeSink) Enabled(level int) bool {
	return t.primary.Enabled(level) || t.secondary.Enabled(level)
}

func (t *teeSink) Info(level int, msg string, keysAndValues ...interface{}) {
	if t.primary.Enabled(level) {
		t.primary.Info(level, msg, keysAndValues...)
	}
	if t.secondary.Enabled(level) {
		t.secondary.Info(level, msg, keysAndValues...)
	}
}

func (t *teeSink) Error(err error, msg string, keysAndValues ...interface{}) {
	t.primary.Error(err, msg, keysAndValues...)
	t.secondary.Error(err, msg, keysAndValues...)
}

func (t *teeSink) WithValues(keysAndValues ...interface{}) logr.LogSink {
	return &teeSink{
		primary:   t.primary.WithValues(keysAndValues...),
		secondary: t.secondary.WithValues(keysAndValues...),
	}
}

func (t *teeSink) WithName(name string) logr.LogSink {
	return &teeSink{
		primary:   t.primary.WithName(name),
		secondary: t.secondary.WithName(name),
	}
}

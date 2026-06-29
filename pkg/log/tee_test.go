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
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
)

func newCapture() (logr.Logger, *[]string) {
	msgs := &[]string{}
	l := funcr.New(func(_, args string) { *msgs = append(*msgs, args) }, funcr.Options{})
	return l, msgs
}

func TestTeeSinkDuplicates(t *testing.T) {
	p, pMsgs := newCapture()
	s, sMsgs := newCapture()
	logger := logr.New(&teeSink{primary: p.GetSink(), secondary: s.GetSink()})

	logger.Info("hello", "k", "v")
	logger.Error(fmt.Errorf("err"), "fail")

	if len(*pMsgs) != 2 || len(*sMsgs) != 2 {
		t.Fatalf("expected 2 msgs each, got primary=%d secondary=%d", len(*pMsgs), len(*sMsgs))
	}
}

func TestTeeSinkWithNameAndValues(t *testing.T) {
	p, _ := newCapture()
	s, sMsgs := newCapture()

	logger := logr.New(&teeSink{primary: p.GetSink(), secondary: s.GetSink()})
	logger.WithName("pg").WithValues("cluster", "c1").Info("msg")

	if len(*sMsgs) != 1 {
		t.Fatal("expected 1 message")
	}
}

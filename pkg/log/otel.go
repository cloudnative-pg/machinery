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
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	uberzap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// openTelemetryScope is the instrumentation scope of the records emitted
// by an unnamed logger
const openTelemetryScope = "github.com/cloudnative-pg/machinery/pkg/log"

var (
	// openTelemetryProvider is the provider exporting the log records,
	// nil when the OpenTelemetry export is disabled
	openTelemetryProvider *sdklog.LoggerProvider

	// openTelemetryForwarder writes only to the OpenTelemetry export, nil
	// when the OpenTelemetry export is disabled
	openTelemetryForwarder zapcore.Core
)

// WithOpenTelemetry also exports the log records to an OpenTelemetry
// collector via OTLP, when the OTEL_EXPORTER_OTLP_ENDPOINT or the
// OTEL_EXPORTER_OTLP_LOGS_ENDPOINT environment variable is set. The
// exporter is configured with the standard OTEL_EXPORTER_OTLP_* variables,
// and OTEL_EXPORTER_OTLP_LOGS_PROTOCOL or OTEL_EXPORTER_OTLP_PROTOCOL
// selects between "http/protobuf", the default, and "grpc". The records
// are exported at the configured log level, in the background: when the
// collector cannot keep up, they are dropped instead of blocking the
// caller. Call ShutdownOpenTelemetry before the process exits.
func WithOpenTelemetry() ConfigureOption {
	return func(o *configureOptions) {
		o.openTelemetry = true
	}
}

// ShutdownOpenTelemetry exports the log records that are still queued and
// stops the OpenTelemetry export, whose later records are only written to
// the log destination. It is a no-op when the export is disabled or already
// stopped.
func ShutdownOpenTelemetry(ctx context.Context) error {
	if openTelemetryProvider == nil {
		return nil
	}

	return openTelemetryProvider.Shutdown(ctx)
}

// ForwardToOpenTelemetry exports, without writing it on the log
// destination, a record that another process wrote in JSON with the
// default keys of the encoder, such as a process spawned by this one whose
// output is already written there. The record keeps its time, level,
// logger name and fields, while a line that is not a JSON object is
// exported as an info message. It is a no-op when the OpenTelemetry export
// is disabled.
func ForwardToOpenTelemetry(line []byte) {
	if openTelemetryForwarder == nil {
		return
	}

	ent, fields := parseJSONRecord(line)
	if ce := openTelemetryForwarder.Check(ent, nil); ce != nil {
		ce.Write(fields...)
	}
}

// parseJSONRecord decodes a record written by the JSON encoder with its
// default keys, mapping the error to the attribute of the semantic
// conventions, as the OpenTelemetry bridge does for the records of this
// process
func parseJSONRecord(line []byte) (zapcore.Entry, []zapcore.Field) {
	ent := zapcore.Entry{Time: time.Now(), Level: InfoLevel}

	var record map[string]any
	decoder := json.NewDecoder(bytes.NewReader(line))
	decoder.UseNumber()
	if err := decoder.Decode(&record); err != nil {
		ent.Message = string(line)
		return ent, nil
	}

	if ts, ok := record["ts"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			ent.Time = parsed
		}
	}
	level, _ := record["level"].(string)
	ent.Level = getLogLevel(level)
	ent.LoggerName, _ = record["logger"].(string)
	ent.Message, _ = record["msg"].(string)
	ent.Stack, _ = record["stacktrace"].(string)
	for _, key := range []string{"ts", "level", "logger", "msg", "stacktrace"} {
		delete(record, key)
	}
	if err, ok := record["error"]; ok {
		delete(record, "error")
		record["exception.message"] = err
	}

	fields := make([]zapcore.Field, 0, len(record))
	for _, key := range slices.Sorted(maps.Keys(record)) {
		value := record[key]
		if number, ok := value.(json.Number); ok {
			if integer, err := number.Int64(); err == nil {
				value = integer
			} else {
				value, _ = number.Float64()
			}
		}
		fields = append(fields, uberzap.Any(key, value))
	}

	return ent, fields
}

func isOpenTelemetryConfigured() bool {
	return os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT") != "" ||
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != ""
}

// setupOpenTelemetry creates the provider exporting the log records, and
// returns the zap option that tees the records to it
func setupOpenTelemetry(level zapcore.Level) (zap.Opts, error) {
	exporter, err := newOpenTelemetryExporter()
	if err != nil {
		return nil, err
	}

	openTelemetryProvider = sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)
	core := &openTelemetryCore{
		LevelEnabler: level,
		core:         otelzap.NewCore(openTelemetryScope, otelzap.WithLoggerProvider(openTelemetryProvider)),
	}
	openTelemetryForwarder = core

	return func(in *zap.Options) {
		in.ZapOpts = append(in.ZapOpts, uberzap.WrapCore(func(c zapcore.Core) zapcore.Core {
			// The log destination is written first: its encoder replaces the
			// Kubernetes objects in the shared fields with their reference,
			// which is what this core then exports
			return zapcore.NewTee(c, core)
		}))
	}, nil
}

func newOpenTelemetryExporter() (sdklog.Exporter, error) {
	protocol := os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL")
	if protocol == "" {
		protocol = os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	}

	switch protocol {
	case "", "http/protobuf":
		return otlploghttp.New(context.Background())
	case "grpc":
		return otlploggrpc.New(context.Background())
	default:
		return nil, fmt.Errorf("unsupported OTLP protocol %q", protocol)
	}
}

// openTelemetryCore adapts the OpenTelemetry zap bridge to this package:
// it filters the records at the configured log level, it maps the custom
// debug and trace levels to the zap debug level, which the bridge
// understands, and it encodes reflected values in JSON and back, exporting
// structs as maps, with the same keys they have on the log destination.
// It adds itself in Check, because a tee writes every core with the entry
// of the first one, so the level must be mapped in Write.
type openTelemetryCore struct {
	zapcore.LevelEnabler
	core zapcore.Core
}

func (c *openTelemetryCore) With(fields []zapcore.Field) zapcore.Core {
	return &openTelemetryCore{LevelEnabler: c.LevelEnabler, core: c.core.With(toJSONFields(fields))}
}

func (c *openTelemetryCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}

	return ce
}

func (c *openTelemetryCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	ent.Level = max(ent.Level, zapcore.DebugLevel)
	return c.core.Write(ent, toJSONFields(fields))
}

func (c *openTelemetryCore) Sync() error {
	return c.core.Sync()
}

// toJSONFields returns a copy of the fields where every reflected value is
// replaced by its JSON representation, leaving the other cores of the tee
// untouched
func toJSONFields(fields []zapcore.Field) []zapcore.Field {
	result := make([]zapcore.Field, len(fields))
	for i, field := range fields {
		result[i] = field
		if field.Type != zapcore.ReflectType {
			continue
		}

		content, err := json.Marshal(field.Interface)
		if err != nil {
			continue
		}

		var value any
		if err := json.Unmarshal(content, &value); err != nil {
			continue
		}
		result[i] = uberzap.Any(field.Key, value)
	}

	return result
}

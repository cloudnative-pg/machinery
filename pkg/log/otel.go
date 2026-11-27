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
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	uberzap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/credentials"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	kjson "sigs.k8s.io/json"
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
// OTEL_EXPORTER_OTLP_LOGS_ENDPOINT environment variable is set, unless
// OTEL_LOGS_EXPORTER is "none" or OTEL_SDK_DISABLED is "true". The exporter
// is configured with the standard OTEL_EXPORTER_OTLP_* variables, and
// OTEL_EXPORTER_OTLP_LOGS_PROTOCOL or OTEL_EXPORTER_OTLP_PROTOCOL
// selects between "http/protobuf", the default, and "grpc". The records
// are exported at the configured log level, in the background: when the
// collector cannot keep up, they are dropped instead of blocking the
// caller. Call ShutdownOpenTelemetry before the process exits. Configure
// the logging once: a later call replaces the export without stopping the
// previous one, and is not safe to run while records are being forwarded.
//
// It also replaces the global error handler of OpenTelemetry, which covers
// all its signals, with one that writes the errors only to the log
// destination. A caller that wants its own handler sets it after
// ConfigureLogging, and must not log through this package.
func WithOpenTelemetry() ConfigureOption {
	return func(o *configureOptions) {
		o.openTelemetry = true
	}
}

// WithOpenTelemetryLoggerProvider also exports the log records through the
// given provider, whatever the environment, for a process that configures
// OpenTelemetry for its other signals and wants the log records to share
// their resource and exporter selection. The records are exported at the
// configured log level. The caller owns the provider and shuts it down
// before the process exits, as ShutdownOpenTelemetry ignores it. This
// option takes precedence over WithOpenTelemetry, and replaces the global
// error handler of OpenTelemetry in the same way.
func WithOpenTelemetryLoggerProvider(provider otellog.LoggerProvider) ConfigureOption {
	return func(o *configureOptions) {
		o.openTelemetryLoggerProvider = provider
	}
}

// ShutdownOpenTelemetry exports the log records that are still queued and
// stops the OpenTelemetry export, whose later records are only written to
// the log destination. It is a no-op when the export is disabled or already
// stopped, or when the provider was passed with
// WithOpenTelemetryLoggerProvider.
func ShutdownOpenTelemetry(ctx context.Context) error {
	if openTelemetryProvider == nil {
		return nil
	}

	return openTelemetryProvider.Shutdown(ctx)
}

// ForwardToOpenTelemetry exports, without writing it on the log
// destination, a record that another process wrote in JSON, such as a
// process spawned by this one whose output is already written there. The
// process is expected to use the same time and level keys as this one, as
// set by GetFieldsRemapFlags. The record keeps its time, level,
// logger name and fields, while a line that is not a JSON object is
// exported as an info message. The line is parsed whole and queued as is,
// so the caller bounds its size. Records below the configured level are
// dropped, and times must be RFC3339. The process must not export on its
// own, or its records are exported twice. It is a no-op when the
// OpenTelemetry export is disabled.
func ForwardToOpenTelemetry(line []byte) {
	if openTelemetryForwarder == nil {
		return
	}

	ent, fields := parseJSONRecord(line)
	if ce := openTelemetryForwarder.Check(ent, nil); ce != nil {
		ce.Write(fields...)
	}
}

// parseJSONRecord decodes a record written by the JSON encoder, mapping the
// error to the attribute of the semantic conventions, as the OpenTelemetry
// bridge does for the records of this process
func parseJSONRecord(line []byte) (zapcore.Entry, []zapcore.Field) {
	ent := zapcore.Entry{Time: time.Now(), Level: InfoLevel}

	var record map[string]any
	// the JSON literal null decodes without error into a nil map
	if err := kjson.UnmarshalCaseSensitivePreserveInts(line, &record); err != nil || record == nil {
		ent.Message = string(line)
		return ent, nil
	}

	timeKey, levelKey := "ts", "level"
	if logfieldsRemap.TimeKey != "" {
		timeKey = logfieldsRemap.TimeKey
	}
	if logfieldsRemap.LevelKey != "" {
		levelKey = logfieldsRemap.LevelKey
	}

	if ts, ok := record[timeKey].(string); ok {
		if parsed, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			ent.Time = parsed
		}
	}
	level, _ := record[levelKey].(string)
	ent.Level = getLogLevel(level)
	ent.LoggerName, _ = record["logger"].(string)
	ent.Message, _ = record["msg"].(string)
	ent.Stack, _ = record["stacktrace"].(string)
	for _, key := range []string{timeKey, levelKey, "logger", "msg", "stacktrace"} {
		delete(record, key)
	}
	if err, ok := record["error"]; ok {
		delete(record, "error")
		record["exception.message"] = err
	}

	fields := make([]zapcore.Field, 0, len(record))
	for _, key := range slices.Sorted(maps.Keys(record)) {
		fields = append(fields, uberzap.Any(key, record[key]))
	}

	return ent, fields
}

// isOpenTelemetryConfigured tells whether the environment asks to export the
// logs
func isOpenTelemetryConfigured() bool {
	// OTEL_SDK_DISABLED turns off every signal, while "none" is the standard
	// way to turn off the export of the logs only, even when the endpoint is
	// set for the other signals
	if strings.EqualFold(os.Getenv("OTEL_SDK_DISABLED"), "true") ||
		strings.EqualFold(os.Getenv("OTEL_LOGS_EXPORTER"), "none") {
		return false
	}

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

	return teeToOpenTelemetry(openTelemetryProvider, level), nil
}

// teeToOpenTelemetry returns the zap option that tees the records to the
// given provider
func teeToOpenTelemetry(provider otellog.LoggerProvider, level zapcore.Level) zap.Opts {
	core := &openTelemetryCore{
		LevelEnabler: level,
		core:         otelzap.NewCore(openTelemetryScope, otelzap.WithLoggerProvider(provider)),
	}
	openTelemetryForwarder = core

	return func(in *zap.Options) {
		in.ZapOpts = append(in.ZapOpts, uberzap.WrapCore(func(c zapcore.Core) zapcore.Core {
			// The log destination is written first: its encoder replaces the
			// Kubernetes objects in the shared fields with their reference,
			// which is what this core then exports
			handleOpenTelemetryErrors(c)

			return zapcore.NewTee(c, core)
		}))
	}
}

// handleOpenTelemetryErrors writes the errors of the OpenTelemetry SDK, such
// as a failed export, only to the log destination. The default handler
// prints them with the standard logger, which is redirected to the tee: a
// failed export would be exported, fail and be logged again.
func handleOpenTelemetryErrors(destination zapcore.Core) {
	logger := uberzap.New(destination)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.Error("OpenTelemetry error", uberzap.Error(err))
	}))
}

// newOpenTelemetryExporter creates the OTLP exporter selected by the
// protocol variables
func newOpenTelemetryExporter() (sdklog.Exporter, error) {
	protocol := os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL")
	if protocol == "" {
		protocol = os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	}

	tlsConfig, err := newOpenTelemetryTLSConfig()
	if err != nil {
		return nil, err
	}

	switch protocol {
	case "", "http/protobuf":
		var opts []otlploghttp.Option
		if tlsConfig != nil {
			opts = append(opts, otlploghttp.WithTLSClientConfig(tlsConfig))
		}
		return otlploghttp.New(context.Background(), opts...)
	case "grpc":
		var opts []otlploggrpc.Option
		if tlsConfig != nil {
			opts = append(opts, otlploggrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)))
		}
		return otlploggrpc.New(context.Background(), opts...)
	default:
		return nil, fmt.Errorf("unsupported OTLP protocol %q", protocol)
	}
}

// newOpenTelemetryTLSConfig returns the TLS configuration of an exporter with
// a client certificate, nil when the environment sets none. The SDK reads
// the client certificate once, while this configuration reads it again at
// every new connection, so that a rotated certificate is used without
// restarting the process. The server certificate authority is read once.
func newOpenTelemetryTLSConfig() (*tls.Config, error) {
	certFile, keyFile := firstOpenTelemetryClientKeyPair()
	if certFile == "" {
		return nil, nil
	}

	config := &tls.Config{
		GetClientCertificate: func(request *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
			if err != nil {
				return nil, err
			}

			// A certificate the server does not accept is not presented,
			// as the TLS library does with the one the SDK reads: a
			// server with an optional client authentication keeps
			// accepting the connection
			if request.SupportsCertificate(&certificate) != nil {
				return &tls.Certificate{}, nil
			}

			return &certificate, nil
		},
	}

	for _, name := range []string{
		"OTEL_EXPORTER_OTLP_LOGS_CERTIFICATE",
		"OTEL_EXPORTER_OTLP_CERTIFICATE",
	} {
		path := os.Getenv(name)
		if path == "" {
			continue
		}

		content, err := os.ReadFile(path) // #nosec G304 G703 -- the path comes from the OTLP standard variable
		if err != nil {
			return nil, err
		}
		config.RootCAs = x509.NewCertPool()
		if !config.RootCAs.AppendCertsFromPEM(content) {
			return nil, errors.New("certificate not added from " + path)
		}
		break
	}

	return config, nil
}

// firstOpenTelemetryClientKeyPair returns the files of the client
// certificate and key, taking the logs variables before the generic ones as
// the SDK does, and empty when no pair is complete
func firstOpenTelemetryClientKeyPair() (string, string) {
	for _, names := range [][2]string{
		{"OTEL_EXPORTER_OTLP_LOGS_CLIENT_CERTIFICATE", "OTEL_EXPORTER_OTLP_LOGS_CLIENT_KEY"},
		{"OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE", "OTEL_EXPORTER_OTLP_CLIENT_KEY"},
	} {
		certFile, keyFile := os.Getenv(names[0]), os.Getenv(names[1])
		if certFile != "" && keyFile != "" {
			return certFile, keyFile
		}
	}

	return "", ""
}

// openTelemetryCore adapts the OpenTelemetry zap bridge to this package:
// it filters the records at the configured log level, it maps the custom
// debug and trace levels to the zap debug level, which the bridge
// understands, and it encodes reflected values in JSON and back, exporting
// structs as maps, with the same keys they have on the log destination.
// A tee writes every core with the entry of the first one, so the level
// cannot be mapped in Check: Write maps it.
type openTelemetryCore struct {
	zapcore.LevelEnabler
	core zapcore.Core
}

// With returns a core with the fields added, as JSON values
func (c *openTelemetryCore) With(fields []zapcore.Field) zapcore.Core {
	return &openTelemetryCore{LevelEnabler: c.LevelEnabler, core: c.core.With(toJSONFields(fields))}
}

// Check adds the core to the entry when its level is enabled
func (c *openTelemetryCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}

	return ce
}

// Write exports the record, mapping the custom levels to the debug one
func (c *openTelemetryCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	ent.Level = max(ent.Level, zapcore.DebugLevel)
	return c.core.Write(ent, toJSONFields(fields))
}

// Sync does nothing, as the bridge buffers nothing: ShutdownOpenTelemetry
// flushes the queue
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
		if err := kjson.UnmarshalCaseSensitivePreserveInts(content, &value); err != nil {
			continue
		}
		result[i] = uberzap.Any(field.Key, value)
	}

	return result
}

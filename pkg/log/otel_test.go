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
	"encoding/pem"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"time"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// exportedRecord is a log record received by the OTLP test receiver
type exportedRecord struct {
	scope      string
	severity   logspb.SeverityNumber
	body       string
	attributes map[string]*commonpb.AnyValue
}

// otlpReceiver collects the log records exported to it
type otlpReceiver struct {
	collogspb.UnimplementedLogsServiceServer

	mu             sync.Mutex
	records        []exportedRecord
	requests       int
	authorizations []string
}

func (r *otlpReceiver) Export(
	_ context.Context,
	req *collogspb.ExportLogsServiceRequest,
) (*collogspb.ExportLogsServiceResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, resourceLogs := range req.GetResourceLogs() {
		for _, scopeLogs := range resourceLogs.GetScopeLogs() {
			for _, record := range scopeLogs.GetLogRecords() {
				attributes := make(map[string]*commonpb.AnyValue)
				for _, kv := range record.GetAttributes() {
					attributes[kv.GetKey()] = kv.GetValue()
				}
				r.records = append(r.records, exportedRecord{
					scope:      scopeLogs.GetScope().GetName(),
					severity:   record.GetSeverityNumber(),
					body:       record.GetBody().GetStringValue(),
					attributes: attributes,
				})
			}
		}
	}

	return &collogspb.ExportLogsServiceResponse{}, nil
}

func (r *otlpReceiver) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	r.requests++
	r.authorizations = append(r.authorizations, req.Header.Get("Authorization"))
	r.mu.Unlock()

	content, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var request collogspb.ExportLogsServiceRequest
	if err := proto.Unmarshal(content, &request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	response, _ := r.Export(req.Context(), &request)
	body, _ := proto.Marshal(response)
	w.Header().Set("Content-Type", "application/x-protobuf")
	_, _ = w.Write(body)
}

// withBody returns the received records having the given body
func (r *otlpReceiver) withBody(body string) []exportedRecord {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []exportedRecord
	for _, record := range r.records {
		if record.body == body {
			result = append(result, record)
		}
	}
	return result
}

// startHTTPReceiver points the OTLP/HTTP exporter at a new receiver
func startHTTPReceiver() *otlpReceiver {
	receiver := &otlpReceiver{}
	server := httptest.NewServer(receiver)
	DeferCleanup(server.Close)
	GinkgoT().Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", server.URL)
	return receiver
}

// startGRPCReceiver points the OTLP/gRPC exporter at a new receiver
func startGRPCReceiver() *otlpReceiver {
	receiver := &otlpReceiver{}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	Expect(err).ToNot(HaveOccurred())
	server := grpc.NewServer()
	collogspb.RegisterLogsServiceServer(server, receiver)
	go func() {
		_ = server.Serve(listener)
	}()
	DeferCleanup(server.Stop)
	GinkgoT().Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
	GinkgoT().Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+listener.Addr().String())
	return receiver
}

func shutdownOpenTelemetry() {
	GinkgoHelper()
	Expect(ShutdownOpenTelemetry(context.Background())).To(Succeed())
}

// testRecord mimics a structured record logged as a value, like the
// PostgreSQL log records
type testRecord struct {
	Message  string `json:"message,omitempty"`
	Severity string `json:"error_severity,omitempty"`
	Query    string `json:"query,omitempty"`
}

var _ = Describe("ConfigureLogging with the OpenTelemetry export", func() {
	var dest string

	BeforeEach(func() {
		dest = filepath.Join(GinkgoT().TempDir(), "log")
		GinkgoT().Setenv("POD_NAME", "cluster-example-1")
	})

	AfterEach(func() {
		logDestination = ""
		if openTelemetryProvider != nil {
			Expect(openTelemetryProvider.Shutdown(context.Background())).To(Succeed())
		}
	})

	It("exports the records via OTLP/HTTP with their severity and structured values", func() {
		receiver := startHTTPReceiver()
		configureTestLogging(dest, nil, WithDisabledSampling(), WithOpenTelemetry())

		WithName("postgres").
			WithValues("source", testRecord{Message: "from WithValues"}).
			Info("record", "record", &testRecord{Message: "division by zero", Severity: "ERROR"})
		Debug("debug-marker")
		Error(errors.New("boom"), "error-marker")
		shutdownOpenTelemetry()

		records := receiver.withBody("record")
		Expect(records).To(HaveLen(1))
		Expect(records[0].scope).To(Equal("postgres"))
		Expect(records[0].severity).To(Equal(logspb.SeverityNumber_SEVERITY_NUMBER_INFO))
		Expect(records[0].attributes["logging_pod"].GetStringValue()).To(Equal("cluster-example-1"))

		fields := make(map[string]string)
		for _, kv := range records[0].attributes["record"].GetKvlistValue().GetValues() {
			fields[kv.GetKey()] = kv.GetValue().GetStringValue()
		}
		Expect(fields).To(Equal(map[string]string{
			"message":        "division by zero",
			"error_severity": "ERROR",
		}))
		source := records[0].attributes["source"].GetKvlistValue().GetValues()
		Expect(source).To(HaveLen(1))
		Expect(source[0].GetKey()).To(Equal("message"))
		Expect(source[0].GetValue().GetStringValue()).To(Equal("from WithValues"))

		Expect(receiver.withBody("debug-marker")).To(BeEmpty(),
			"records below the configured log level must not be exported")
		errorRecords := receiver.withBody("error-marker")
		Expect(errorRecords).To(HaveLen(1))
		Expect(errorRecords[0].severity).To(Equal(logspb.SeverityNumber_SEVERITY_NUMBER_ERROR))

		Expect(destLines(dest, `"msg":"record"`)).To(HaveLen(1),
			"the log destination still gets the records")
	})

	It("exports a Kubernetes object as the reference written on the log destination", func() {
		receiver := startHTTPReceiver()
		configureTestLogging(dest, nil, WithDisabledSampling(), WithOpenTelemetry())

		Info("object-marker", "secret", &metav1.PartialObjectMetadata{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "app-credentials",
				Namespace:   "default",
				Annotations: map[string]string{"content": "not to be exported"},
			},
		})
		shutdownOpenTelemetry()

		records := receiver.withBody("object-marker")
		Expect(records).To(HaveLen(1))
		fields := make(map[string]string)
		for _, kv := range records[0].attributes["secret"].GetKvlistValue().GetValues() {
			fields[kv.GetKey()] = kv.GetValue().GetStringValue()
		}
		Expect(fields).To(Equal(map[string]string{"name": "app-credentials", "namespace": "default"}))
	})

	It("configures the exporter and the batches with the standard environment variables", func() {
		receiver := &otlpReceiver{}
		server := httptest.NewTLSServer(receiver)
		DeferCleanup(server.Close)
		caFile := filepath.Join(GinkgoT().TempDir(), "ca.crt")
		Expect(os.WriteFile(caFile,
			pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0o600)).To(Succeed())
		GinkgoT().Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", server.URL+"/v1/logs")
		GinkgoT().Setenv("OTEL_EXPORTER_OTLP_CERTIFICATE", caFile)
		GinkgoT().Setenv("OTEL_EXPORTER_OTLP_HEADERS", "authorization=Bearer%20secret")
		GinkgoT().Setenv("OTEL_BLRP_MAX_EXPORT_BATCH_SIZE", "1")
		configureTestLogging(dest, nil, WithDisabledSampling(), WithOpenTelemetry())

		Info("tls-marker")
		Info("tls-marker")
		shutdownOpenTelemetry()

		Expect(receiver.withBody("tls-marker")).To(HaveLen(2))
		receiver.mu.Lock()
		defer receiver.mu.Unlock()
		Expect(receiver.authorizations).To(ConsistOf("Bearer secret", "Bearer secret"),
			"each record must be sent in its own batch, with the configured header")
	})

	It("prefers the protocol of the logs signal over the generic one", func() {
		receiver := startGRPCReceiver()
		GinkgoT().Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
		GinkgoT().Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
		GinkgoT().Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "grpc")
		GinkgoT().Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
		configureTestLogging(dest, nil, WithDisabledSampling(), WithOpenTelemetry())

		Info("signal-protocol-marker")
		shutdownOpenTelemetry()

		Expect(receiver.withBody("signal-protocol-marker")).To(HaveLen(1))
	})

	It("exports the debug and trace records via OTLP/gRPC with the debug severity", func() {
		receiver := startGRPCReceiver()
		configureTestLogging(dest, []string{"--log-level", "trace"}, WithDisabledSampling(), WithOpenTelemetry())

		Debug("debug-marker")
		Trace("trace-marker")
		shutdownOpenTelemetry()

		for _, body := range []string{"debug-marker", "trace-marker"} {
			records := receiver.withBody(body)
			Expect(records).To(HaveLen(1))
			Expect(records[0].severity).To(Equal(logspb.SeverityNumber_SEVERITY_NUMBER_DEBUG))
		}
	})

	It("exports the same records the sampler lets through to the log destination", func() {
		// burst must exceed the 100 msgs/s initial-pass threshold of the sampler
		// installed by the controller-runtime zap builder
		const burst = 300

		receiver := startHTTPReceiver()
		configureTestLogging(dest, nil, WithOpenTelemetry())

		for range burst {
			Info("sampled-burst-marker")
		}
		shutdownOpenTelemetry()

		written := destLines(dest, `"msg":"sampled-burst-marker"`)
		Expect(len(written)).To(BeNumerically("<", burst))
		Expect(receiver.withBody("sampled-burst-marker")).To(HaveLen(len(written)))
	})

	It("writes the records of OpenTelemetryLogger only to the export", func() {
		receiver := startHTTPReceiver()
		configureTestLogging(dest, nil, WithDisabledSampling(), WithOpenTelemetry())

		OpenTelemetryLogger().WithName("wal-archive").Info("export-only-marker", "walName", "000000010000000000000001")
		shutdownOpenTelemetry()

		records := receiver.withBody("export-only-marker")
		Expect(records).To(HaveLen(1))
		Expect(records[0].scope).To(Equal("wal-archive"))
		Expect(records[0].attributes["walName"].GetStringValue()).To(Equal("000000010000000000000001"))
		Expect(destLines(dest, "export-only-marker")).To(BeEmpty())
	})

	It("stops the export at the first shutdown, so that a stuck collector delays the exit only once", func() {
		release := make(chan struct{})
		server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			<-release
		}))
		DeferCleanup(server.Close)
		DeferCleanup(func() { close(release) })
		GinkgoT().Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", server.URL)
		configureTestLogging(dest, nil, WithDisabledSampling(), WithOpenTelemetry())

		Info("stuck-marker")
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		Expect(ShutdownOpenTelemetry(ctx)).To(HaveOccurred())

		Info("after-shutdown-marker")
		start := time.Now()
		Expect(ShutdownOpenTelemetry(context.Background())).To(Succeed())
		Expect(time.Since(start)).To(BeNumerically("<", 100*time.Millisecond))
		Expect(destLines(dest, "after-shutdown-marker")).To(HaveLen(1))
	})

	It("does not export anything without the option, even with an endpoint", func() {
		receiver := startHTTPReceiver()
		configureTestLogging(dest, nil, WithDisabledSampling())

		Info("not-exported-marker")
		shutdownOpenTelemetry()

		Expect(receiver.withBody("not-exported-marker")).To(BeEmpty())
		Expect(OpenTelemetryLogger().Enabled()).To(BeFalse())
		Expect(destLines(dest, "not-exported-marker")).To(HaveLen(1))
	})

	It("keeps logging to the log destination with an unsupported protocol", func() {
		receiver := startHTTPReceiver()
		GinkgoT().Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/json")
		configureTestLogging(dest, nil, WithDisabledSampling(), WithOpenTelemetry())

		Info("unsupported-protocol-marker")
		shutdownOpenTelemetry()

		Expect(receiver.withBody("unsupported-protocol-marker")).To(BeEmpty())
		Expect(OpenTelemetryLogger().Enabled()).To(BeFalse())
		Expect(destLines(dest, "unsupported-protocol-marker")).To(HaveLen(1))
	})
})

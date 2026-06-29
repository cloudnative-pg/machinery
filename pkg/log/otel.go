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
	"fmt"
	"os"

	"github.com/go-logr/logr"
	otellogr "go.opentelemetry.io/contrib/bridges/otellogr"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

var otelLogEndpoint string

// OTelShutdown holds the shutdown function for the OTel log provider.
// Call this during graceful shutdown to flush pending logs.
var OTelShutdown = func() {}

func setupOTelLogger(existing logr.Logger) logr.Logger {
	endpoint := otelLogEndpoint
	if endpoint == "" {
		endpoint = os.Getenv("CNPG_LOG_OTEL_ENDPOINT")
	}
	if endpoint == "" {
		return existing
	}

	ctx := context.Background()
	exporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint(endpoint),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "otel log bridge: failed to create exporter: %v\n", err)
		return existing
	}

	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)

	OTelShutdown = func() {
		_ = provider.Shutdown(context.Background())
	}

	otelSink := otellogr.NewLogSink("cnpg",
		otellogr.WithLoggerProvider(provider),
	)

	return logr.New(&teeSink{primary: existing.GetSink(), secondary: otelSink})
}

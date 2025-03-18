// SPDX-License-Identifier: MIT

package tracer

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

var tracer trace.Tracer

// Init makes the global tracer and connects to the traces collector (gRPC).
//
// Returns closer that closes connection and shuts down tracer provider.
func Init(ctx context.Context, appName, version string, opts ...Option) (func(context.Context) error, error) {
	if tracer != nil {
		return nil, errors.New("tracer already initialized")
	}

	options := buildOptions(opts)

	if options.IsNoop() {
		tracer = noop.NewTracerProvider().Tracer("")
		return func(_ context.Context) error {
			return nil
		}, nil
	}

	exporter, closer, err := makeGrpcExporter(ctx, options)
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exporter),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(appName),
			semconv.ServiceVersion(version),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tracer = otel.Tracer("")

	return func(ctx context.Context) error {
		var errs []error
		if err := tp.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown tracer provider: %w", err))
		}
		if closer != nil {
			if err := closer(); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	}, nil
}

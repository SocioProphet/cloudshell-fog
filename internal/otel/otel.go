// Package otel initialises the OpenTelemetry SDK with stdout exporters.
package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Providers holds references to the global tracer and meter after setup.
type Providers struct {
	Tracer   trace.Tracer
	Meter    metric.Meter
	Shutdown func(context.Context) error
}

// Setup initialises the OpenTelemetry SDK with stdout exporters and registers global
// providers. Call Shutdown on program exit to flush and clean up.
func Setup(ctx context.Context, serviceName string) (*Providers, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(attribute.String("service.name", serviceName)),
	)
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}

	traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("create stdout trace exporter: %w", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, fmt.Errorf("create stdout metric exporter: %w", err)
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	shutdown := func(ctx context.Context) error {
		if err := tp.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown tracer: %w", err)
		}
		return mp.Shutdown(ctx)
	}

	return &Providers{
		Tracer:   otel.Tracer(serviceName),
		Meter:    otel.Meter(serviceName),
		Shutdown: shutdown,
	}, nil
}

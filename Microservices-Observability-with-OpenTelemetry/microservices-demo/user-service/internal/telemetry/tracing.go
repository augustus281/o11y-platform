package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

const (
	UserService = "user-service"
	Version     = "1.0.0"
)

// InitTracing initializes the tracing system.
// Set OTEL_EXPORTER_OTLP_ENDPOINT env var to the collector endpoint (e.g., http://localhost:4318)
func InitTracing(ctx context.Context) (func(ctx context.Context) error, error) {
	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("error to new exporter, %w", err)
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(UserService),
			semconv.ServiceVersion(Version),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("error to new resource, %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	slog.Info("Tracing initialized!")

	// return tp.Shutdown to the application to flush all pending traces and stop background workers gracefully when it exists
	return tp.Shutdown, nil
}

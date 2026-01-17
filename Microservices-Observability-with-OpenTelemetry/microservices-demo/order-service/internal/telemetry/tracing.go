package telemetry

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// InitTracing initializes OpenTelemetry tracing for order-service
func InitTracing(ctx context.Context) (func(context.Context) error, error) {
	// OTLP endpoint (Collector)
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
	if endpoint == "" {
		endpoint = "otel-collector.observability.svc.cluster.local:4318"
	}

	// OTLP HTTP exporter
	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(), // use HTTP
	)
	if err != nil {
		return nil, err
	}

	// Resource = service identity
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName("order-service"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	// TracerProvider = tracing engine
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Register globally
	otel.SetTracerProvider(tp)

	// Return shutdown for graceful exit
	return tp.Shutdown, nil
}

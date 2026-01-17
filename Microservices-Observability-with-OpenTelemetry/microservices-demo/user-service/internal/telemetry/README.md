# TRACING

Initializing distributed tracing for the service and exporting spans to OpenTelemetry Collector via OTLP HTTP

## High-level Flow

```text
Load config
   ↓
Create OTLP exporter
   ↓
Define service metadata (Resource)
   ↓
Create TracerProvider
   ↓
Register it globally
   ↓
Return shutdown function
```
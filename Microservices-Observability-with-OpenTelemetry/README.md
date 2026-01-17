# Microservices Observability with OpenTelemetry

Implement comprehensive observability for microservices using OpenTelemetry for distributed tracing, metrics, and logging across multiple services.

## Deploy Observability Stack Infrastructure

### Prerequisites

```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.4/cert-manager.yaml

# Add resource for certificate
kubectl apply -f cert_manager.yaml
```


### Set up Jaeger

Set up Jaeger for tracing, Prometheus for metrics, and configure the foundational observability infrastructure.

```bash
# Create observability namespace
kubectl create namespace observability

# Deploy Jaeger using Operator
kubectl create -f https://github.com/jaegertracing/jaeger-operator/releases/download/v1.39.0/jaeger-operator.yaml -n observability

# Wait for operator to be ready
kubectl wait --for=condition=available deployment jaeger-operator -n observability --timeout=300s

# Deploy OpenTelemetry Collector
kubectl apply -f otel_collector.yaml
```

## Create Instrumented Microservices

Build sample microservices with OpenTelemetry instrumentation for automatic tracing and metrics collection.

#!/bin/bash

export OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://localhost:9090/api/v1/otlp/v1/metrics
export OTEL_TRACES_EXPORTER=none
export OTEL_LOGS_EXPORTER=none
export OTEL_METRIC_EXPORT_INTERVAL=15000
export OTEL_SERVICE_NAME="crud-service"

set -e

docker run --rm -d --name=prometheus -p 9090:9090 \
    -v C:/Users/aidyn/Desktop/crud_project/prometheus.yml:/etc/prometheus/prometheus.yml \
    prom/prometheus
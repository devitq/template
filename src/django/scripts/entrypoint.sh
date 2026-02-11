#!/bin/sh

set -e

# Run migrations if requested
if [ "$RUN_MIGRATIONS" = "true" ]; then
    echo "Running migrations..."
    python manage.py migrate --noinput
fi

# Collect static files if requested
if [ "$COLLECT_STATIC" = "true" ]; then
    echo "Collecting static files..."
    python manage.py collectstatic --noinput
fi

# Start the application
if [ "$OTEL_ENABLED" = "true" ]; then
    echo "Starting with OpenTelemetry instrumentation..."
    export OTEL_PYTHON_LOGGING_AUTO_INSTRUMENTATION_ENABLED="true"
    export OTEL_TRACES_EXPORTER="${OTEL_TRACES_EXPORTER:-otlp}"
    export OTEL_METRICS_EXPORTER="${OTEL_METRICS_EXPORTER:-otlp}"
    export OTEL_LOGS_EXPORTER="${OTEL_LOGS_EXPORTER:-otlp}"
    export OTEL_EXPORTER_OTLP_ENDPOINT="${OTEL_EXPORTER_OTLP_ENDPOINT:-http://otel-collector:4317}"

    exec opentelemetry-instrument \
        --service_name "${OTEL_SERVICE_NAME:-backend-django}" \
        gunicorn config.asgi:application \
        --workers "${GUNICORN_WORKERS:-4}" \
        --worker-class "${GUNICORN_WORKER_CLASS:-uvicorn_worker.UvicornWorker}" \
        --bind "${GUNICORN_BIND:-0.0.0.0:8080}" \
        --access-logfile "${GUNICORN_ACCESS_LOG:--}" \
        --error-logfile "${GUNICORN_ERROR_LOG:--}"
else
    echo "Starting without OpenTelemetry instrumentation..."
    exec gunicorn config.asgi:application \
        --workers "${GUNICORN_WORKERS:-4}" \
        --worker-class "${GUNICORN_WORKER_CLASS:-uvicorn_worker.UvicornWorker}" \
        --bind "${GUNICORN_BIND:-0.0.0.0:8080}" \
        --access-logfile "${GUNICORN_ACCESS_LOG:--}" \
        --error-logfile "${GUNICORN_ERROR_LOG:--}"
fi

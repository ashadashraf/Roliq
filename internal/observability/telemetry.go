package observability

import (
	"context"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

var (
	httpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "roliq_http_requests_total", Help: "Completed HTTP requests."}, []string{"method", "route", "status"})
	httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "roliq_http_request_duration_seconds", Help: "HTTP request latency.", Buckets: prometheus.DefBuckets}, []string{"method", "route"})
)

func init() { prometheus.MustRegister(httpRequests, httpDuration) }

func InitTracing(ctx context.Context, serviceName, endpoint string) (func(context.Context) error, error) {
	if endpoint == "" {
		return func(context.Context) error { return nil }, nil
	}
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(endpoint))
	if err != nil {
		return nil, err
	}
	provider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter), sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(serviceName))))
	otel.SetTracerProvider(provider)
	return provider.Shutdown, nil
}

func MetricsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		started := time.Now()
		err := next(c)
		route := c.Path()
		if route == "" {
			route = "unmatched"
		}
		httpRequests.WithLabelValues(c.Request().Method, route, strconv.Itoa(c.Response().Status)).Inc()
		httpDuration.WithLabelValues(c.Request().Method, route).Observe(time.Since(started).Seconds())
		return err
	}
}

func TracingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := otel.Tracer("github.com/roliq/roliq/internal/httpapi").Start(c.Request().Context(), c.Request().Method+" "+c.Path())
		defer span.End()
		c.SetRequest(c.Request().WithContext(ctx))
		err := next(c)
		route := c.Path()
		if route == "" {
			route = "unmatched"
		}
		span.SetAttributes(attribute.String("http.request.method", c.Request().Method), attribute.String("http.route", route), attribute.Int("http.response.status_code", c.Response().Status))
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "request failed")
		}
		return err
	}
}

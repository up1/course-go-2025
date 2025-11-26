package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	echomid "github.com/labstack/echo/v4/middleware"

	// Tracing

	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	// Metrics
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// Swagger
	_ "demo/docs" // Import generated swagger docs

	echoSwagger "github.com/swaggo/echo-swagger"

	"demo/handlers"
)

// @title Product API
// @version 1.0
// @description This is a sample server for a Product API.
// @host localhost:8080
// @BasePath /
func main() {
	// --- 1. Structured Logging Setup (slog) ---
	// Set the default logger to JSON output
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger.Info("Starting API service")

	// --- 2. OpenTelemetry Setup (Tracing) ---
	// In a real app, you would set up a TraceProvider, MetricProvider, and Exporter here.
	// For simplicity, we skip the exporter configuration, but we use the API.
	// `shutdownTracing := setupTelemetry(logger)`
	// `defer shutdownTracing()`
	shutdownTracing := setupTelemetry(logger)
	defer shutdownTracing()

	e := echo.New()
	e.Logger.SetOutput(os.Stdout) // Pipe Echo's internal logger to stdout
	e.HideBanner = true

	// --- 3. Middleware ---
	// Distributed Tracing: Must be near the top
	e.Use(otelecho.Middleware("product-service"))

	// Custom Request Logging with slog and tracing context
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract trace and span IDs from OpenTelemetry context
			spanCtx := c.Request().Context()
			span := trace.SpanFromContext(spanCtx)
			traceID := span.SpanContext().TraceID().String()

			c.Set("logger", logger.With(slog.String("traceID", traceID)))

			// Log start of request (or use a modified Echo Logger middleware)
			logger.Info("Incoming request",
				slog.String("method", c.Request().Method),
				slog.String("path", c.Request().URL.Path),
				slog.String("traceID", traceID),
			)
			return next(c)
		}
	})

	// Recovery middleware (to prevent crashes)
	e.Use(echomid.Recover())

	// --- 4. Product Handler and Route ---
	h := &handlers.ProductHandler{Logger: logger}
	e.GET("/product/:id", h.GetProductByID)

	// --- 5. Prometheus Metrics Endpoint ---
	// Expose metrics on a separate, dedicated port/group for scraping
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// --- 6. Auto-Generate Swagger/OpenAPI Document ---
	// Requires running 'swag init' first to generate 'docs/'
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// --- 7. Start Server with Graceful Shutdown ---
	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", slog.Any("error", err))
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt) // CTRL+C
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with a 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", slog.Any("error", err))
	}

	logger.Info("Server stopped.")
}

func setupTelemetry(logger *slog.Logger) func() {
	// Initialize OpenTelemetry Tracing
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tracerProvider)

	// Initialize OpenTelemetry Metrics
	metricProvider := metric.NewMeterProvider()
	otel.SetMeterProvider(metricProvider)

	otel.SetTracerProvider(tracerProvider)

	return func() {
		// Shutdown procedures
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			logger.Error("Failed to shutdown tracer provider", slog.Any("error", err))
		}
	}
}

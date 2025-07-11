package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"handler/function"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
)

var (
	acceptingConnections int32
)

const defaultTimeout = 10 * time.Second

func getExporters(ctx context.Context) ([]sdktrace.SpanExporter, error) {
	var exporters []sdktrace.SpanExporter

	exporterTypes := os.Getenv("OTEL_TRACES_EXPORTER")
	if exporterTypes == "" {
		exporterTypes = "stdout"
	}

	insecureValue := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_INSECURE")
	insecure, err := strconv.ParseBool(insecureValue)
	if err != nil {
		insecure = false
	}

	for _, exp := range strings.Split(exporterTypes, ",") {
		switch exp {
		case "console":
			exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
			if err != nil {
				return nil, err
			}

			exporters = append(exporters, exporter)
		case "otlp":
			endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
			if endpoint == "" {
				endpoint = "0.0.0.0:4317"
			}

			opts := []otlptracegrpc.Option{
				otlptracegrpc.WithEndpoint(endpoint),
				otlptracegrpc.WithDialOption(grpc.WithBlock()),
			}

			if insecure {
				opts = append(opts, otlptracegrpc.WithInsecure())

			}

			exporter, err := otlptracegrpc.New(ctx, opts...)
			if err != nil {
				return nil, err
			}

			exporters = append(exporters, exporter)
		default:
			fmt.Printf("unknown OTEL exporter type: %s", exp)
		}
	}

	return exporters, nil
}

func newTracerProvider(exporters []sdktrace.SpanExporter) *sdktrace.TracerProvider {
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		if functionName, ok := os.LookupEnv("OPENFAAS_NAME"); ok {
			serviceName = functionName

			// If we are running in a kubernetes cluster, use the namespace in the service name
			namespace := ""
			data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
			if err == nil {
				namespace = string(data)
			}

			if len(namespace) > 0 {
				serviceName = fmt.Sprintf("%s.%s", serviceName, namespace)
			}
		}
	}

	if serviceName == "" {
		serviceName = "unknown-function"
	}

	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)

	if err != nil {
		panic(err)
	}

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(r),
	}

	for _, exporter := range exporters {
		processor := sdktrace.NewBatchSpanProcessor(exporter)
		opts = append(opts, sdktrace.WithSpanProcessor(processor))
	}

	return sdktrace.NewTracerProvider(opts...)
}

func main() {
	ctx := context.Background()
	exporters, err := getExporters(ctx)
	if err != nil {
		log.Fatalf("failed to initialize exporters: %v", err)
	}

	// Create a new tracer provider with a batch span processor and the given exporters.
	tp := newTracerProvider(exporters)
	otel.SetTracerProvider(tp)

	// Handle shutdown properly so nothing leaks.
	defer func() { _ = tp.Shutdown(ctx) }()

	readTimeout := parseIntOrDurationValue(os.Getenv("read_timeout"), defaultTimeout)
	writeTimeout := parseIntOrDurationValue(os.Getenv("write_timeout"), defaultTimeout)
	healthInterval := parseIntOrDurationValue(os.Getenv("healthcheck_interval"), writeTimeout)

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", 8082),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	http.Handle("/", otelhttp.NewHandler(http.HandlerFunc(function.Handle), "Invoke"))

	listenUntilShutdown(s, healthInterval, writeTimeout)
}

func listenUntilShutdown(s *http.Server, shutdownTimeout time.Duration, writeTimeout time.Duration) {
	idleConnsClosed := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM)

		<-sig

		log.Printf("[entrypoint] SIGTERM: no connections in: %s", shutdownTimeout.String())
		<-time.Tick(shutdownTimeout)

		ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			log.Printf("[entrypoint] Error in Shutdown: %v", err)
		}

		log.Printf("[entrypoint] Exiting.")

		close(idleConnsClosed)
	}()

	// Run the HTTP server in a separate go-routine.
	go func() {
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("[entrypoint] Error ListenAndServe: %v", err)
			close(idleConnsClosed)
		}
	}()

	atomic.StoreInt32(&acceptingConnections, 1)

	<-idleConnsClosed
}

func parseIntOrDurationValue(val string, fallback time.Duration) time.Duration {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return time.Duration(parsedVal) * time.Second
		}
	}

	duration, durationErr := time.ParseDuration(val)
	if durationErr != nil {
		return fallback
	}
	return duration
}

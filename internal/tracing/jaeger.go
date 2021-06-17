package tracing

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.uber.org/zap"
)


func InitJaegerTracer(serviceName, environment string, logger *zap.SugaredLogger) *sdktrace.TracerProvider {
	exporter, err := jaeger.NewRawExporter(jaeger.WithAgentEndpoint())
	if err != nil {
		logger.With("error", err).Fatal("could not register tracing jaeger exporter")
	}

	res := resource.NewWithAttributes(
		semconv.ServiceNameKey.String(serviceName),
		attribute.String("environment", environment),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}

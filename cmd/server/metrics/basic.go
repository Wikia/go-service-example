package metrics

import (
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	export "go.opentelemetry.io/otel/sdk/export/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv"
	"go.uber.org/zap"
)

var GreetCount metric.Int64Counter

func RegisterMetrics(appName, environment string, logger *zap.SugaredLogger) (*prometheus.Exporter, *controller.Controller) {
	config := prometheus.Config{}
	res := resource.NewWithAttributes(
		semconv.ServiceNameKey.String(appName),
		attribute.String("environment", environment),
	)
	c := controller.New(
		processor.New(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(config.DefaultHistogramBoundaries),
			),
			export.CumulativeExportKindSelector(),
			processor.WithMemory(true),
		),
		controller.WithResource(res),
	)

	exporter, err := prometheus.New(config, c)
	if err != nil {
		logger.With("error", err).Panic("failed to initialize prometheus exporter")
	}
	global.SetMeterProvider(exporter.MeterProvider())

	if err := runtime.Start(
		runtime.WithMinimumReadMemStatsInterval(time.Second),
	); err != nil {
		logger.With("error", err).Panic("could not start the runtime metrics collector")
	}

	m := global.Meter(appName)

	GreetCount = metric.Must(m).NewInt64Counter(
		"greets_total",
		metric.WithDescription("Number of generated greetings"),
	)

	return exporter, c
}
